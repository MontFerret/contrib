package cdp

import (
	"context"

	net "github.com/MontFerret/contrib/modules/web/html/drivers/cdp/network"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type filteredNavigationEventStream struct {
	stream runtime.Stream
	match  func(*net.NavigationEvent) bool
}

func newFilteredNavigationEventStream(
	stream runtime.Stream,
	match func(*net.NavigationEvent) bool,
) runtime.Stream {
	return &filteredNavigationEventStream{
		stream: stream,
		match:  match,
	}
}

func (f *filteredNavigationEventStream) Close() error {
	return f.stream.Close()
}

func (f *filteredNavigationEventStream) Read(ctx context.Context) <-chan runtime.Message {
	out := make(chan runtime.Message)

	go func() {
		defer close(out)

		for evt := range f.stream.Read(ctx) {
			if err := evt.Err(); err != nil {
				out <- runtime.NewErrorMessage(err)
				return
			}

			nav, ok := evt.Value().(*net.NavigationEvent)
			if !ok {
				continue
			}

			if !f.match(nav) {
				continue
			}

			out <- runtime.NewValueMessage(nav)
		}
	}()

	return out
}
