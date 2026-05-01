package cdp

import (
	"context"
	"testing"
	"time"

	netdriver "github.com/MontFerret/contrib/modules/web/html/drivers/cdp/network"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type stubStream struct {
	messages []runtime.Message
	closed   int
}

func (s *stubStream) Close() error {
	s.closed++
	return nil
}

func (s *stubStream) Read(context.Context) <-chan runtime.Message {
	out := make(chan runtime.Message, len(s.messages))
	defer close(out)

	for _, message := range s.messages {
		out <- message
	}

	return out
}

func TestPreparedNavigationEventStream(t *testing.T) {
	raw := &stubStream{
		messages: []runtime.Message{
			runtime.NewValueMessage((&netdriver.NavigationEvent{
				URL:     "https://example.com",
				FrameID: "frame-id",
			}).Copy()),
		},
	}

	prepareCalls := 0
	stream := newPreparedNavigationEventStream(raw, func(_ context.Context, evt *netdriver.NavigationEvent) error {
		prepareCalls++
		if evt.FrameID != "frame-id" {
			t.Fatalf("expected frame-id, got %s", evt.FrameID)
		}

		return nil
	})

	for message := range stream.Read(context.Background()) {
		if err := message.Err(); err != nil {
			t.Fatalf("unexpected stream error: %v", err)
		}
	}

	if prepareCalls != 1 {
		t.Fatalf("expected prepare to run once, got %d", prepareCalls)
	}
}

// blockingStubStream emits one event then blocks until ctx is cancelled,
// mirroring how real upstream streams behave with a slow/idle source.
type blockingStubStream struct {
	first runtime.Message
}

func (b *blockingStubStream) Close() error { return nil }

func (b *blockingStubStream) Read(ctx context.Context) <-chan runtime.Message {
	out := make(chan runtime.Message)

	go func() {
		defer close(out)

		select {
		case <-ctx.Done():
			return
		case out <- b.first:
		}

		<-ctx.Done()
	}()

	return out
}

func TestPreparedNavigationEventStreamExitsOnCtxCancel(t *testing.T) {
	raw := &blockingStubStream{
		first: runtime.NewValueMessage((&netdriver.NavigationEvent{
			URL:     "https://example.com",
			FrameID: "frame-id",
		}).Copy()),
	}

	stream := newPreparedNavigationEventStream(raw, func(context.Context, *netdriver.NavigationEvent) error {
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	ch := stream.Read(ctx)

	// Receive the first message, then stop consuming and cancel.
	select {
	case <-ch:
	case <-time.After(time.Second):
		cancel()
		t.Fatal("timed out waiting for first message")
	}

	cancel()

	// After ctx is cancelled, the fan-out goroutine must exit and close the channel.
	select {
	case _, ok := <-ch:
		if ok {
			// Drain any straggler and re-check.
			for range ch {
				continue
			}
		}
	case <-time.After(time.Second):
		t.Fatal("fan-out goroutine did not exit after ctx cancel")
	}
}

func TestPreparedNavigationEventStreamCloseIsTransportOnly(t *testing.T) {
	raw := &stubStream{}
	prepareCalls := 0

	stream := newPreparedNavigationEventStream(raw, func(context.Context, *netdriver.NavigationEvent) error {
		prepareCalls++
		return nil
	})

	if err := stream.Close(); err != nil {
		t.Fatalf("unexpected close error: %v", err)
	}

	if raw.closed != 1 {
		t.Fatalf("expected raw stream to close once, got %d", raw.closed)
	}

	if prepareCalls != 0 {
		t.Fatalf("expected close to avoid prepare, got %d calls", prepareCalls)
	}
}
