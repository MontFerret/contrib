package cdp

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mafredri/cdp/protocol/browser"
	"github.com/mafredri/cdp/rpcc"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
)

func TestPolicySessionCleanupFailure(t *testing.T) {
	var calls atomic.Int32
	endpoint := newPolicyTestEndpoint(t, func(_ int32, method string, _ json.RawMessage) (any, error) {
		if method == "Target.closeTarget" {
			calls.Add(1)

			return nil, &rpcc.ResponseError{Code: -32000, Message: "cannot close target"}
		}

		return map[string]any{"targetInfos": []any{map[string]any{"targetId": "live-target"}}}, nil
	})
	conn, client := endpoint.client(t)
	policy := newBrowserPolicy([]string{})
	lifetime, invalidate := context.WithCancel(t.Context())
	ctx, cancel := context.WithCancel(t.Context())
	session := &policySession{policy: policy, conn: conn, client: client, targetID: "live-target", ctx: ctx, cancel: cancel, lifetime: lifetime, invalidate: invalidate}
	policy.sessions[session] = struct{}{}
	page := &HTMLPage{policySession: session}
	session.page = page

	err := page.Close()
	if err == nil || !strings.Contains(err.Error(), "live-target") || calls.Load() != 1 {
		t.Fatalf("failed target cleanup: err=%v calls=%d", err, calls.Load())
	}

	if again := page.Close(); again != err {
		t.Fatalf("repeated page close changed its result: %v", again)
	}

	if err := policy.Close(); err == nil || !strings.Contains(err.Error(), "cannot close target") {
		t.Fatalf("driver lost the page cleanup failure: %v", err)
	}
}

func TestPolicySessionCleanupFallback(t *testing.T) {
	for _, test := range []struct {
		name        string
		primaryOK   bool
		fallbackOK  bool
		unavailable bool
		gone        bool
		isolated    bool
	}{
		{name: "primary success", primaryOK: true},
		{name: "fallback success", fallbackOK: true},
		{name: "primary disconnected", unavailable: true, fallbackOK: true},
		{name: "both fail"},
		{name: "target already gone", gone: true},
		{name: "context fallback", isolated: true, fallbackOK: true},
		{name: "context already gone", isolated: true, gone: true},
		{name: "context still present", isolated: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			var primaryCalls, fallbackCalls atomic.Int32
			endpoint := newPolicyTestEndpoint(t, func(id int32, method string, _ json.RawMessage) (any, error) {
				switch method {
				case "Target.closeTarget", "Target.disposeBrowserContext":
					if id == 1 {
						fallbackCalls.Add(1)
						if test.fallbackOK {
							return map[string]any{"success": true}, nil
						}
					} else {
						primaryCalls.Add(1)
						if test.primaryOK {
							return map[string]any{"success": true}, nil
						}
					}

					return nil, &rpcc.ResponseError{Code: -32000, Message: "cleanup refused"}
				case "Target.getTargets":
					targets := []any{}
					if !test.gone {
						targets = append(targets, map[string]any{"targetId": "owned-target"})
					}

					return map[string]any{"targetInfos": targets}, nil
				case "Target.getBrowserContexts":
					contexts := []string{}
					if !test.gone {
						contexts = append(contexts, "owned-context")
					}

					return map[string]any{"browserContextIds": contexts}, nil
				}

				return struct{}{}, nil
			})
			policy, session := newPolicyCleanupSession(t, endpoint)
			if test.isolated {
				session.contextID = browser.ContextID("owned-context")
			}

			if test.unavailable {
				_ = session.conn.Close()
			}

			err := session.Close()
			wantFailure := !test.primaryOK && !test.fallbackOK && !test.gone
			if (err != nil) != wantFailure {
				t.Fatalf("cleanup error = %v", err)
			}

			if test.primaryOK && fallbackCalls.Load() != 0 {
				t.Fatal("successful cleanup attempted a fallback")
			}

			if test.unavailable && primaryCalls.Load() != 0 {
				t.Fatal("cleanup used a disconnected channel")
			}

			if endpoint.opened.Load() != 2 {
				t.Fatal("cleanup opened a replacement connection")
			}

			if closeErr := policy.Close(); (closeErr != nil) != wantFailure {
				t.Fatalf("policy close error = %v", closeErr)
			}
		})
	}
}

func newPolicyCleanupSession(t *testing.T, endpoint *policyTestEndpoint) (*browserPolicy, *policySession) {
	t.Helper()
	policyConn, _ := endpoint.client(t)
	conn, client := endpoint.client(t)
	policy := newBrowserPolicy([]string{})
	policy.conn = policyConn
	lifetime, invalidate := context.WithCancel(t.Context())
	ctx, cancel := context.WithCancel(t.Context())
	session := &policySession{policy: policy, conn: conn, client: client, targetID: "owned-target", ctx: ctx, cancel: cancel, lifetime: lifetime, invalidate: invalidate}
	policy.sessions[session] = struct{}{}
	t.Cleanup(func() { _ = policy.Close() })

	return policy, session
}

func TestPolicySessionCleanupBudgetAndConcurrentClose(t *testing.T) {
	var primary, fallback atomic.Int32
	endpoint := newPolicyTestEndpoint(t, func(id int32, method string, _ json.RawMessage) (any, error) {
		if method == "Target.closeTarget" {
			if id == 2 {
				primary.Add(1)

				return nil, nil
			}

			fallback.Add(1)

			return map[string]any{"success": true}, nil
		}

		return struct{}{}, nil
	})
	policy, session := newPolicyCleanupSession(t, endpoint)
	page := &HTMLPage{policySession: session}
	session.page = page
	var wg sync.WaitGroup
	started := time.Now()
	for index := range 8 {
		wg.Go(func() {
			close := page.Close
			if index%2 == 0 {
				close = policy.Close
			}

			if err := close(); err != nil {
				t.Errorf("concurrent close: %v", err)
			}
		})
	}

	wg.Wait()
	if elapsed := time.Since(started); elapsed > policyCleanupTimeout {
		t.Fatalf("fallback exceeded shared cleanup budget: %s", elapsed)
	}

	if primary.Load() != 1 || fallback.Load() != 1 {
		t.Fatalf("duplicate cleanup: primary=%d fallback=%d", primary.Load(), fallback.Load())
	}

	if err := policy.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestPolicySessionCleanupSharesDeadline(t *testing.T) {
	endpoint := newPolicyTestEndpoint(t, func(_ int32, method string, _ json.RawMessage) (any, error) {
		if method == "Target.closeTarget" {
			return nil, nil
		}

		return struct{}{}, nil
	})
	policy, session := newPolicyCleanupSession(t, endpoint)
	started := time.Now()
	err := session.Close()
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("cleanup deadline = %v", err)
	}

	if elapsed := time.Since(started); elapsed > policyCleanupTimeout+time.Second {
		t.Fatalf("cleanup attempts did not share a deadline: %s", elapsed)
	}

	if err := policy.Close(); err == nil {
		t.Fatal("driver lost timed-out cleanup")
	}
}

func TestPolicySessionPartialInitialization(t *testing.T) {
	for _, stage := range []string{"discovery", "auto attach", "canceled discovery", "canceled discovery with cleanup failure"} {
		t.Run(stage, func(t *testing.T) {
			canceled := strings.HasPrefix(stage, "canceled")
			cleanupFailure := strings.HasSuffix(stage, "cleanup failure")
			var closed atomic.Int32
			discovering := make(chan struct{}, 1)
			endpoint := newPolicyTestEndpoint(t, func(_ int32, method string, _ json.RawMessage) (any, error) {
				switch method {
				case "Target.createTarget":
					return map[string]any{"targetId": "partial-target"}, nil
				case "Target.setDiscoverTargets":
					discovering <- struct{}{}
					if stage == "auto attach" {
						return struct{}{}, nil
					}

					if canceled {
						return nil, nil
					}

					return nil, &rpcc.ResponseError{Code: -32000, Message: "discovery failed"}
				case "Target.autoAttachRelated":
					return nil, &rpcc.ResponseError{Code: -32000, Message: "auto attachment failed"}
				case "Target.closeTarget":
					closed.Add(1)
					if cleanupFailure {
						return nil, &rpcc.ResponseError{Code: -32000, Message: "cleanup refused"}
					}

					return map[string]any{"success": true}, nil
				case "Target.getTargets":
					return map[string]any{"targetInfos": []any{map[string]any{"targetId": "partial-target"}}}, nil
				}

				return struct{}{}, nil
			})
			driver := New(WithAddress(endpoint.server.URL), WithAllowedOrigins())
			defer driver.Close()
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			finished := make(chan error, 1)
			go func() { _, err := driver.Parse(ctx, drivers.ParseParams{}); finished <- err }()
			<-discovering
			if canceled {
				cancel()
			}

			err := <-finished
			if err == nil || (canceled && !errors.Is(err, context.Canceled)) {
				t.Fatalf("initialization failure = %v", err)
			}

			if cleanupFailure {
				if !strings.Contains(err.Error(), "cleanup refused") || driver.Close() == nil || closed.Load() != 2 {
					t.Fatalf("initialization lost cleanup failure: %v, attempts=%d", err, closed.Load())
				}
			} else if closed.Load() != 1 {
				t.Fatalf("partial initialization left its target: closures=%d", closed.Load())
			}
		})
	}
}

func TestBrowserPolicyRetainsTerminalCleanupFailure(t *testing.T) {
	endpoint := newPolicyTestEndpoint(t, func(_ int32, method string, _ json.RawMessage) (any, error) {
		switch method {
		case "Target.closeTarget":
			return nil, &rpcc.ResponseError{Code: -32000, Message: "terminal cleanup failed"}
		case "Target.getTargets":
			return map[string]any{"targetInfos": []any{map[string]any{"targetId": "terminal-target"}}}, nil
		}

		return struct{}{}, nil
	})
	driver := New(WithAddress(endpoint.server.URL), WithAllowedOrigins())
	defer driver.Close()
	session, err := driver.policy.begin(t.Context(), driver)
	if err != nil {
		t.Fatal(err)
	}

	conn, client := endpoint.client(t)
	if err := session.setConnection(conn, client); err != nil {
		t.Fatal(err)
	}

	session.targetID = "terminal-target"
	if err := driver.policy.interceptor.Close(); err != nil {
		t.Fatal(err)
	}

	select {
	case <-driver.policy.watchDone:
	case <-time.After(time.Second):
		t.Fatal("terminal cleanup did not finish")
	}

	err = driver.Close()
	if err == nil || !strings.Contains(err.Error(), "terminal cleanup failed") {
		t.Fatalf("Close lost terminal cleanup failure: %v", err)
	}

	if again := driver.Close(); again != err {
		t.Fatalf("repeated Close changed its result: %v", again)
	}
}
