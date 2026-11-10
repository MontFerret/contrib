package cdp

import (
	"context"
	"encoding/json"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
)

func TestBrowserPolicyWaiterCancellation(t *testing.T) {
	enabling := make(chan struct{}, 1)
	endpoint := newPolicyTestEndpoint(t, func(_ int32, method string, _ json.RawMessage) (any, error) {
		if method == "Fetch.enable" {
			enabling <- struct{}{}

			return nil, nil
		}

		return struct{}{}, nil
	})
	driver := New(WithAddress(endpoint.server.URL), WithAllowedOrigins())
	t.Cleanup(func() { _ = driver.Close() })
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	first := make(chan error, 1)
	go func() { _, err := driver.policy.begin(ctx, driver); first <- err }()
	<-enabling
	waiterCtx, stop := context.WithCancel(t.Context())
	finished := make(chan error, 1)
	go func() { _, err := driver.policy.begin(waiterCtx, driver); finished <- err }()
	stop()

	select {
	case err := <-finished:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("waiter cancellation = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("canceled waiter remained blocked behind initialization")
	}

	if endpoint.opened.Load() != 1 || endpoint.active.Load() != 1 {
		t.Fatal("canceling a waiter changed the active initialization attempt")
	}

	cancel()
	if err := <-first; !errors.Is(err, context.Canceled) {
		t.Fatalf("initiator cancellation = %v", err)
	}
}

func TestBrowserPolicyCloseCancelsInitialization(t *testing.T) {
	enabling := make(chan struct{}, 1)
	var targets atomic.Int32
	endpoint := newPolicyTestEndpoint(t, func(_ int32, method string, _ json.RawMessage) (any, error) {
		if method == "Fetch.enable" {
			enabling <- struct{}{}

			return nil, nil
		}

		targets.Add(1)

		return struct{}{}, nil
	})
	driver := New(WithAddress(endpoint.server.URL), WithAllowedOrigins())
	defer driver.Close()
	opened := make(chan error, 1)
	go func() { _, err := driver.Open(t.Context(), drivers.Params{}); opened <- err }()
	<-enabling
	closed := make(chan error, 1)
	go func() { closed <- driver.Close() }()

	select {
	case err := <-closed:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("Close did not cancel stalled Fetch.enable")
	}

	select {
	case <-endpoint.closed:
	case <-time.After(time.Second):
		t.Fatal("Close retained the startup connection")
	}

	if err := <-opened; err == nil {
		t.Fatal("canceled initialization succeeded")
	}

	if _, err := driver.Parse(t.Context(), drivers.ParseParams{}); err == nil {
		t.Fatal("closed driver accepted Parse")
	}

	if targets.Load() != 0 || endpoint.opened.Load() != 1 || driver.policy.conn != nil {
		t.Fatal("closed initialization published resources or created targets")
	}
}

func TestBrowserPolicyCanceledInitiatorAllowsRetry(t *testing.T) {
	enabling := make(chan int32, 2)
	var firstAttempt atomic.Pointer[policyInitialization]
	endpoint := newPolicyTestEndpoint(t, func(id int32, method string, _ json.RawMessage) (any, error) {
		if method == "Fetch.enable" {
			enabling <- id
			if id == 1 {
				return nil, nil
			}

			select {
			case <-firstAttempt.Load().done:
			default:
				t.Error("retry overlapped incomplete initialization cleanup")
			}
		}

		return struct{}{}, nil
	})
	driver := New(WithAddress(endpoint.server.URL), WithAllowedOrigins())
	t.Cleanup(func() { _ = driver.Close() })
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	first := make(chan error, 1)
	go func() { _, err := driver.policy.begin(ctx, driver); first <- err }()
	<-enabling
	driver.policy.mu.Lock()
	firstAttempt.Store(driver.policy.initializing)
	driver.policy.mu.Unlock()
	second := make(chan error, 1)
	go func() {
		session, err := driver.policy.begin(t.Context(), driver)
		if err == nil {
			err = session.Close()
		}

		second <- err
	}()
	cancel()
	if err := <-first; !errors.Is(err, context.Canceled) {
		t.Fatalf("initiator error = %v", err)
	}

	select {
	case err := <-second:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("live waiter did not retry canceled initialization")
	}

	if endpoint.opened.Load() != 2 {
		t.Fatalf("initialization connections = %d, want 2", endpoint.opened.Load())
	}
}
