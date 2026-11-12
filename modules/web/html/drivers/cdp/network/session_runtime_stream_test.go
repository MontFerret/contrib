package network

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/mafredri/cdp/rpcc"

	cdpsession "github.com/MontFerret/contrib/modules/web/html/drivers/cdp/session"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestSessionSubscriptionBuffersBeforeRead(t *testing.T) {
	sessions := newSubscriptionTestSessions(&cdpsession.Client{ID: "root"})
	raw := newSubscriptionTestStream()
	opened := false
	stream, err := newSessionRuntimeStream(t.Context(), sessions, "", func(context.Context, *cdpsession.Client) (runtime.Stream, error) {
		opened = true

		return raw, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	defer stream.Close()

	if !opened {
		t.Fatal("subscription returned before opening the current session")
	}

	raw.messages <- runtime.NewValueMessage(runtime.String("before read"))
	select {
	case message := <-stream.Read(t.Context()):
		if message.Value() != runtime.String("before read") {
			t.Fatalf("lost early event: %v", message)
		}
	case <-time.After(time.Second):
		t.Fatal("early event was not retained")
	}
}

func TestSessionSubscriptionCloseBeforeRead(t *testing.T) {
	for _, action := range []string{"close", "cancel", "read cancellation"} {
		t.Run(action, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			sessions := newSubscriptionTestSessions(&cdpsession.Client{ID: "root"})
			raw := newSubscriptionTestStream()
			stream, err := newSessionRuntimeStream(ctx, sessions, "", func(context.Context, *cdpsession.Client) (runtime.Stream, error) {
				return raw, nil
			})
			if err != nil {
				t.Fatal(err)
			}

			switch action {
			case "close":
				closeSubscription(t, stream)
			case "cancel":
				cancel()
			case "read cancellation":
				readCtx, stopRead := context.WithCancel(t.Context())
				stream.Read(readCtx)
				stopRead()
			}

			select {
			case <-stream.(*sessionRuntimeStream).done:
			case <-time.After(time.Second):
				t.Fatal("subscription cleanup stalled")
			}

			closeSubscription(t, stream)
			if raw.closes.Load() != 1 || sessions.listenerCount() != 0 {
				t.Fatalf("cleanup: closes=%d listeners=%d", raw.closes.Load(), sessions.listenerCount())
			}
		})
	}
}

func TestSessionSubscriptionSetupFailure(t *testing.T) {
	want := errors.New("cannot open attached stream")
	sessions := newSubscriptionTestSessions(&cdpsession.Client{ID: "a-root"}, &cdpsession.Client{ID: "b-bad"})
	raw := newSubscriptionTestStream()

	stream, err := newSessionRuntimeStream(t.Context(), sessions, "", func(_ context.Context, client *cdpsession.Client) (runtime.Stream, error) {
		if client.ID == "b-bad" {
			return nil, want
		}

		return raw, nil
	})
	if stream != nil || !errors.Is(err, want) {
		t.Fatalf("got %v, %v; want setup error", stream, err)
	}

	if raw.closes.Load() != 1 || sessions.listenerCount() != 0 {
		t.Fatalf("partial setup leaked: closes=%d listeners=%d", raw.closes.Load(), sessions.listenerCount())
	}
}

func TestSessionSubscriptionAttachmentAndDetachment(t *testing.T) {
	sessions := newSubscriptionTestSessions()
	raw := newSubscriptionTestStream()
	stream, err := newSessionRuntimeStream(t.Context(), sessions, "", func(context.Context, *cdpsession.Client) (runtime.Stream, error) {
		return raw, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	defer stream.Close()

	client := &cdpsession.Client{ID: "child"}
	sessions.emit(cdpsession.EventAttached, client)
	raw.messages <- runtime.NewValueMessage(runtime.String("attached before read"))
	select {
	case message := <-stream.Read(t.Context()):
		if message.Value() != runtime.String("attached before read") {
			t.Fatalf("unexpected attached event: %v", message)
		}
	case <-time.After(time.Second):
		t.Fatal("attached session was not subscribed")
	}

	sessions.emit(cdpsession.EventDetached, client)
	closeSubscription(t, stream)
	if raw.closes.Load() != 1 {
		t.Fatalf("detached stream closed %d times", raw.closes.Load())
	}
}

func TestSessionSubscriptionAttachmentFailureBeforeRead(t *testing.T) {
	sessions := newSubscriptionTestSessions()
	want := errors.New("attachment stream failed")
	stream, err := newSessionRuntimeStream(t.Context(), sessions, "", func(context.Context, *cdpsession.Client) (runtime.Stream, error) {
		return nil, want
	})
	if err != nil {
		t.Fatal(err)
	}

	defer stream.Close()

	finished := make(chan struct{})
	go func() {
		sessions.emit(cdpsession.EventAttached, &cdpsession.Client{ID: "child"})
		close(finished)
	}()
	select {
	case <-finished:
	case <-time.After(time.Second):
		t.Fatal("session callback waited for the first read")
	}

	select {
	case message := <-stream.Read(t.Context()):
		if !errors.Is(message.Err(), want) {
			t.Fatalf("attachment error = %v", message.Err())
		}
	case <-time.After(time.Second):
		t.Fatal("attachment setup error was lost")
	}
}

func TestSessionSubscriptionCloseDuringAttachment(t *testing.T) {
	sessions := newSubscriptionTestSessions()
	started := make(chan struct{})
	stream, err := newSessionRuntimeStream(t.Context(), sessions, "", func(ctx context.Context, _ *cdpsession.Client) (runtime.Stream, error) {
		close(started)
		<-ctx.Done()

		return nil, ctx.Err()
	})
	if err != nil {
		t.Fatal(err)
	}

	var callbacks sync.WaitGroup
	callbacks.Add(1)
	go func() {
		defer callbacks.Done()
		sessions.emit(cdpsession.EventAttached, &cdpsession.Client{ID: "child"})
	}()
	<-started
	closeSubscription(t, stream)
	callbacks.Wait()
	if sessions.listenerCount() != 0 {
		t.Fatal("session listener survived close")
	}
}

func closeSubscription(t *testing.T, stream runtime.Stream) {
	t.Helper()
	done := make(chan error, 1)
	go func() { done <- stream.Close() }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("close stalled")
	}
}

func TestSessionSubscriptionSkipsDetachedSnapshotEntry(t *testing.T) {
	client := &cdpsession.Client{ID: "detached"}
	sessions := newSubscriptionTestSessions(client)
	sessions.afterSnapshot = func() { sessions.emit(cdpsession.EventDetached, client) }
	opened := false
	stream, err := newSessionRuntimeStream(t.Context(), sessions, "", func(context.Context, *cdpsession.Client) (runtime.Stream, error) {
		opened = true

		return newSubscriptionTestStream(), nil
	})
	if err != nil {
		t.Fatal(err)
	}

	defer stream.Close()

	if opened {
		t.Fatal("reopened a detached session from the initial snapshot")
	}
}

func TestRootLossDuringSubscriptionSetup(t *testing.T) {
	sessions := newSubscriptionTestSessions(&cdpsession.Client{ID: "root"})
	var closePage context.CancelCauseFunc
	sessions.pageCtx, closePage = context.WithCancelCause(t.Context())
	stream, err := newSessionRuntimeStream(t.Context(), sessions, "", func(ctx context.Context, _ *cdpsession.Client) (runtime.Stream, error) {
		closePage(rpcc.ErrConnClosing)
		<-ctx.Done()

		return nil, ctx.Err()
	})
	if stream != nil || !errors.Is(err, rpcc.ErrConnClosing) {
		t.Fatalf("subscription during root loss: %v, %v", stream, err)
	}
}
