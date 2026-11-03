package network

import (
	"context"
	"runtime"
	"testing"

	"github.com/mafredri/cdp"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
)

func BenchmarkNetworkSubscription(b *testing.B) {
	ctx := context.Background()
	observer := newStartedTestNetworkObserver(ctx, &cdp.Client{})
	b.ReportAllocs()

	for b.Loop() {
		stream, err := observer.Subscribe(ctx, drivers.NetworkRequestStartedEvent, nil)
		if err != nil {
			b.Fatal(err)
		}

		messages := stream.Read(ctx)
		for {
			observer.mu.Lock()
			ready := len(observer.subscribers) == 1
			observer.mu.Unlock()
			if ready {
				break
			}

			runtime.Gosched()
		}

		observer.emit(networkEvent{name: drivers.NetworkRequestStartedEvent})
		if message := <-messages; message.Err() != nil {
			b.Fatal(message.Err())
		}

		if err := stream.Close(); err != nil {
			b.Fatal(err)
		}

		for range messages {
		}
	}
}
