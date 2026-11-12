package session

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mafredri/cdp/protocol/target"
	"github.com/mafredri/cdp/rpcc"
)

func receive[T any](t *testing.T, ch <-chan T) T {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for CDP event")
		var zero T

		return zero
	}
}

func TestRouting(t *testing.T) {
	p := newProtocolPeer(t)
	var stream rpcc.Stream
	attached := make(chan error, 1)
	p.manager.AddListener(func(e Event) {
		if e.Kind == EventAttached && e.Client.ID == "worker" {
			var err error
			stream, err = rpcc.NewStream(t.Context(), "Test.event", e.Client.Conn)
			attached <- err
		}
	})
	p.attach("worker", "worker-target", "worker")
	p.event("worker", "Test.event", 1)
	enable := p.request("Runtime.enable")
	p.event("worker", "Test.event", 2)
	p.reply(enable, struct{}{})
	p.reply(p.request("Network.enable"), struct{}{})

	if err := receive(t, attached); err != nil {
		t.Fatal(err)
	}

	defer stream.Close()
	for want := 1; want <= 2; want++ {
		var value int

		if err := stream.RecvMsg(&value); err != nil || value != want {
			t.Fatalf("event %d: %d, %v", want, value, err)
		}
	}

	p.attach("worker", "worker-target", "worker")
	p.attach("duplicate", "worker-target", "worker")
	p.reply(p.request("Target.detachFromTarget"), struct{}{})

	worker, ok := p.manager.ClientByTarget("worker-target")
	if !ok || worker.ID != "worker" {
		t.Fatalf("duplicate replaced active worker: %v", worker)
	}

	for range 2 {
		completed := make(chan error, 1)
		go func() { completed <- p.manager.Root().CDP.Runtime.Enable(t.Context()) }()
		p.reply(p.request("Runtime.enable"), struct{}{})

		if err := receive(t, completed); err != nil {
			t.Fatal(err)
		}
	}

	type result struct {
		value string
		err   error
	}

	calls := map[target.SessionID]chan result{}
	for id, conn := range map[target.SessionID]*rpcc.Conn{"root": p.manager.Root().Conn, "worker": worker.Conn, "": p.manager.browserConn} {
		ch := make(chan result, 1)
		calls[id] = ch
		go func() {
			var value string
			err := rpcc.Invoke(t.Context(), "Test.echo", nil, &value, conn)
			ch <- result{value, err}
		}()
	}

	requests := make([]protocolRequest, 3)
	for i := range requests {
		requests[i] = p.request("Test.echo")
	}

	for i := len(requests) - 1; i >= 0; i-- {
		r := requests[i]
		if r.SessionID == "worker" {
			p.send(map[string]any{"id": r.ID, "error": rpcc.ResponseError{Code: -32000, Message: "worker unavailable"}})
		} else {
			p.reply(r, string(r.SessionID))
		}
	}

	for id, ch := range calls {
		got := receive(t, ch)

		if id == "worker" {
			var rpcErr *rpcc.ResponseError

			if !errors.As(got.err, &rpcErr) || rpcErr.Message != "worker unavailable" {
				t.Fatalf("worker error = %v", got.err)
			}
		} else if got.err != nil || got.value != string(id) {
			t.Fatalf("session %q got %+v", id, got)
		}
	}
}

func TestFailedChildAndDetachDuringSetup(t *testing.T) {
	p := newProtocolPeer(t)
	events := make(chan Event, 8)
	p.manager.AddListener(func(e Event) { events <- e })
	p.attach("frame", "frame-target", "iframe")
	page := p.request("Page.enable")
	p.send(map[string]any{"id": page.ID, "sessionId": "frame", "error": rpcc.ResponseError{Code: -32000, Message: "cannot enable page"}})
	failure := receive(t, events)

	_, err := p.manager.FrameSessions()
	if failure.Kind != EventFailed || failure.Client.ID != "frame" || failure.Operation != "initialize" || !errors.Is(err, failure.Err) {
		t.Fatalf("failure = %+v, enumeration error = %v", failure, err)
	}

	p.reply(p.request("Target.detachFromTarget"), struct{}{})
	p.event("", "Target.targetDestroyed", target.DestroyedReply{TargetID: "frame-target"})

	p.attach("short", "short-target", "worker")
	_ = p.request("Runtime.enable")
	p.event("", "Target.detachedFromTarget", target.DetachedFromTargetReply{SessionID: "short"})
	// A later live session is a barrier for the preceding lifecycle events.
	p.attach("live", "live-target", "worker")
	p.reply(p.request("Runtime.enable"), struct{}{})
	p.reply(p.request("Network.enable"), struct{}{})

	event := receive(t, events)
	if event.Kind != EventAttached || event.Client.ID != "live" || p.manager.Root() == nil {
		t.Fatalf("healthy sessions: %+v", event)
	}

	if len(p.manager.Snapshot()) != 2 {
		t.Fatalf("leaked partial session: %v", p.manager.Snapshot())
	}

	if _, err := p.manager.FrameSessions(); err != nil {
		t.Fatal(err)
	}
}

func TestPendingCallsDuringShutdown(t *testing.T) {
	for _, mode := range []string{"cancel", "close", "disconnect"} {
		t.Run(mode, func(t *testing.T) {
			p := newProtocolPeer(t)
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			done := make(chan error, 1)
			root := p.manager.Root().CDP
			go func() { done <- root.Runtime.Enable(ctx) }()
			r := p.request("Runtime.enable")
			switch mode {
			case "cancel":
				cancel()

				if err := receive(t, done); !errors.Is(err, context.Canceled) {
					t.Fatal(err)
				}

				p.reply(r, struct{}{})

				return
			case "close":
				var wg sync.WaitGroup
				for range 8 {
					wg.Go(func() {
						if err := p.manager.Close(); err != nil {
							t.Error(err)
						}
					})
				}

				wg.Wait()
			case "disconnect":
				_ = p.remote.Close()
			}

			if err := receive(t, done); err == nil {
				t.Fatal("pending command survived shutdown")
			}
		})
	}
}

func TestWorkerWithoutNetworkDomain(t *testing.T) {
	p := newProtocolPeer(t)
	attached := make(chan Event, 1)
	p.manager.AddListener(func(e Event) {
		if e.Kind == EventAttached {
			attached <- e
		}
	})
	p.attach("worker", "worker-target", "worker")
	p.reply(p.request("Runtime.enable"), struct{}{})
	network := p.request("Network.enable")
	p.send(map[string]any{"id": network.ID, "sessionId": "worker", "error": rpcc.ResponseError{Code: -32601, Message: "Method not found"}})

	worker := receive(t, attached).Client
	if worker.SupportsNetwork() || worker.ID != "worker" {
		t.Fatalf("worker capabilities: %+v", worker)
	}
}

func TestRejectsConflictingResponseSession(t *testing.T) {
	p := newProtocolPeer(t)
	finished := make(chan error, 1)
	root := p.manager.Root()
	go func() { finished <- root.CDP.Runtime.Enable(t.Context()) }()
	call := p.request("Runtime.enable")
	p.send(map[string]any{"id": call.ID, "sessionId": "different", "result": struct{}{}})

	if err := receive(t, finished); err == nil {
		t.Fatal("conflicting response completed the root call")
	}

	select {
	case <-p.manager.Context().Done():
	case <-time.After(time.Second):
		t.Fatal("invalid routing left the connection active")
	}
}

func TestCloseDiscoversPopupAfterOpenerClosesItself(t *testing.T) {
	p := newProtocolPeer(t)
	finished := make(chan error, 1)
	go func() { finished <- p.manager.ClosePageTargets(t.Context(), nil) }()

	root, popup := target.ID("page"), target.ID("popup")
	unrelated := target.Info{TargetID: "sibling", Type: "page"}
	p.reply(p.request("Target.closeTarget"), struct{}{})
	p.reply(p.request("Target.getTargets"), target.GetTargetsReply{TargetInfos: []target.Info{
		{TargetID: popup, Type: "page", OpenerID: &root}, unrelated,
	}})

	closePopup := p.request("Target.closeTarget")
	if closePopup.Params.TargetID != popup {
		t.Fatalf("closing %q, want %q", closePopup.Params.TargetID, popup)
	}

	p.send(map[string]any{"id": closePopup.ID, "error": rpcc.ResponseError{Code: -32000, Message: "target already closed"}})
	p.reply(p.request("Target.getTargets"), target.GetTargetsReply{TargetInfos: []target.Info{
		{TargetID: "nested", Type: "page", OpenerID: &popup}, unrelated,
	}})

	closeNested := p.request("Target.closeTarget")
	if closeNested.Params.TargetID != "nested" {
		t.Fatalf("closing %q, want nested popup", closeNested.Params.TargetID)
	}

	p.reply(closeNested, struct{}{})
	p.reply(p.request("Target.getTargets"), target.GetTargetsReply{TargetInfos: []target.Info{unrelated}})

	if err := receive(t, finished); err != nil {
		t.Fatal(err)
	}
}

func TestChildSetupLimits(t *testing.T) {
	for _, mode := range []string{"deadline", "events", "bytes"} {
		t.Run(mode, func(t *testing.T) {
			p := newProtocolPeer(t)
			events := make(chan Event, 8)
			p.manager.AddListener(func(e Event) { events <- e })
			p.attach("healthy", "healthy-target", "worker")
			p.reply(p.request("Runtime.enable"), struct{}{})
			p.reply(p.request("Network.enable"), struct{}{})
			sibling := receive(t, events).Client

			if mode == "deadline" {
				router := p.manager.targets.router
				router.mu.Lock()
				p.manager.targets.setupTimeout = 100 * time.Millisecond
				router.mu.Unlock()
			}

			p.attach("slow", "slow-frame", "iframe")
			enable := p.request("Page.enable")
			command := func() {
				done := make(chan error, 1)
				go func() { done <- sibling.CDP.Runtime.Enable(t.Context()) }()

				request := p.request("Runtime.enable")
				if request.SessionID != sibling.ID {
					t.Fatal(request)
				}

				p.reply(request, struct{}{})
				if err := receive(t, done); err != nil {
					t.Fatal(err)
				}
			}

			command()

			want := errSetupOverflow
			switch mode {
			case "events":
				for range maxSetupEvents + 1 {
					p.event("slow", "Test.early", 1)
				}
			case "bytes":
				p.event("slow", "Test.early", strings.Repeat("x", maxSetupBytes))
			case "deadline":
				want = context.DeadlineExceeded
			}

			failure := receive(t, events)
			if failure.Kind != EventFailed || failure.Client.ID != "slow" || !errors.Is(failure.Err, want) || !errors.Is(p.manager.FrameError("slow-frame"), want) {
				t.Fatalf("failure: %+v", failure)
			}

			detach := p.request("Target.detachFromTarget")
			p.event("", "Target.detachedFromTarget", target.DetachedFromTargetReply{SessionID: "slow"})
			p.reply(detach, struct{}{})
			p.reply(enable, struct{}{})
			command()
			select {
			case extra := <-events:
				t.Fatalf("additional lifecycle event: %+v", extra)
			default:
			}
		})
	}
}
