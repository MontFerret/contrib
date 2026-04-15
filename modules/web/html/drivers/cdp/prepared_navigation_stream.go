package cdp

import (
	"context"

	net "github.com/MontFerret/contrib/modules/web/html/drivers/cdp/network"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type preparedNavigationEventStream struct {
	stream  runtime.Stream
	prepare func(context.Context, *net.NavigationEvent) error
}

func newPreparedNavigationEventStream(
	stream runtime.Stream,
	prepare func(context.Context, *net.NavigationEvent) error,
) runtime.Stream {
	return &preparedNavigationEventStream{
		stream:  stream,
		prepare: prepare,
	}
}

func (p *preparedNavigationEventStream) Close() error {
	return p.stream.Close()
}

func (p *preparedNavigationEventStream) Read(ctx context.Context) <-chan runtime.Message {
	out := make(chan runtime.Message)

	go func() {
		defer close(out)

		for evt := range p.stream.Read(ctx) {
			if err := evt.Err(); err != nil {
				out <- runtime.NewErrorMessage(err)
				return
			}

			nav, ok := evt.Value().(*net.NavigationEvent)
			if !ok {
				continue
			}

			if err := p.prepare(ctx, nav); err != nil {
				out <- runtime.NewErrorMessage(err)
				return
			}

			out <- runtime.NewValueMessage(nav)
		}
	}()

	return out
}
