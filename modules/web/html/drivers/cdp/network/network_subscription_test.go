package network

import (
	"context"
	"testing"
	"time"

	"github.com/mafredri/cdp"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
)

func TestNetworkSubscriptionBuffersBeforeRead(t *testing.T) {
	observer := newStartedTestNetworkObserver(t.Context(), &cdp.Client{})
	defer observer.cancel()
	stream, err := observer.Subscribe(t.Context(), drivers.NetworkRequestStartedEvent, nil)
	if err != nil {
		t.Fatal(err)
	}

	defer stream.Close()

	observer.emit(networkEvent{name: drivers.NetworkRequestStartedEvent, url: "https://example.test/early"})
	payload := readNetworkObject(t, stream.Read(t.Context()))
	assertStringField(t, payload, "url", "https://example.test/early")
}

func TestNetworkSubscriptionCleanupBeforeRead(t *testing.T) {
	for _, name := range []string{drivers.NetworkRequestStartedEvent, drivers.NetworkIdleEvent} {
		for _, action := range []string{"close", "cancel"} {
			t.Run(name+"/"+action, func(t *testing.T) {
				ctx, cancel := context.WithCancel(t.Context())
				defer cancel()
				observer := newStartedTestNetworkObserver(ctx, &cdp.Client{})
				defer observer.cancel()
				stream, err := observer.Subscribe(ctx, name, nil)
				if err != nil {
					t.Fatal(err)
				}

				observer.mu.Lock()
				if len(observer.subscribers) != 1 {
					observer.mu.Unlock()
					t.Fatal("subscription is not active before Read")
				}

				var done <-chan struct{}
				for _, subscriber := range observer.subscribers {
					done = subscriber.done
				}

				observer.mu.Unlock()

				if action == "cancel" {
					cancel()
				} else if err := stream.Close(); err != nil {
					t.Fatal(err)
				}

				select {
				case <-done:
				case <-time.After(time.Second):
					t.Fatal("unread subscription survived cleanup")
				}

				if err := stream.Close(); err != nil {
					t.Fatal(err)
				}

				observer.mu.Lock()
				remaining := len(observer.subscribers)
				observer.mu.Unlock()
				if remaining != 0 {
					t.Fatalf("%d subscribers remain", remaining)
				}

				select {
				case _, ok := <-stream.Read(t.Context()):
					if ok {
						t.Fatal("closed subscription produced a message")
					}
				case <-time.After(time.Second):
					t.Fatal("read after close stalled")
				}
			})
		}
	}
}
