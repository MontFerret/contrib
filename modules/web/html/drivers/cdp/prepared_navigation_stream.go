package cdp

import (
	"context"
	"errors"

	"github.com/mafredri/cdp/rpcc"

	net "github.com/MontFerret/contrib/modules/web/html/drivers/cdp/network"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type preparedNavigationEventStream struct {
	stream   runtime.Stream
	lifetime context.Context
	cancel   context.CancelFunc
	page     context.Context
	opts     net.WaitEventOptions
	prepare  func(context.Context, *net.NavigationEvent) error
}

func newPreparedNavigationEventStream(
	ctx, page context.Context,
	stream runtime.Stream,
	opts net.WaitEventOptions,
	prepare func(context.Context, *net.NavigationEvent) error,
) runtime.Stream {
	lifetime, cancel := context.WithCancel(ctx)

	return &preparedNavigationEventStream{
		stream:   stream,
		lifetime: lifetime,
		cancel:   cancel,
		page:     page,
		opts:     opts,
		prepare:  prepare,
	}
}

func (s *preparedNavigationEventStream) Close() error {
	s.cancel()

	return s.stream.Close()
}

func (s *preparedNavigationEventStream) Read(ctx context.Context) <-chan runtime.Message {
	out := make(chan runtime.Message, 1)
	go func() {
		defer close(out)
		read, cancel := context.WithCancel(ctx)
		defer cancel()
		stop := context.AfterFunc(s.lifetime, cancel)
		defer stop()
		stopPage := context.AfterFunc(s.page, cancel)
		defer stopPage()

		for msg := range s.stream.Read(read) {
			if msg.Err() == nil {
				nav, ok := msg.Value().(*net.NavigationEvent)
				if !ok || (s.opts.FrameID != "" && nav.FrameID != s.opts.FrameID) {
					continue
				}

				if err := s.prepare(read, nav); err != nil {
					msg = runtime.NewErrorMessage(err)
				} else if !matchNavigationEvent(nav, s.opts) {
					continue
				}
			}

			if s.page.Err() != nil {
				break
			}

			select {
			case <-read.Done():
			case out <- msg:
			}

			if read.Err() != nil || msg.Err() != nil {
				break
			}
		}

		// Replace a queued navigation when its page closes, even for a slow reader.
		if err := context.Cause(s.page); errors.Is(err, rpcc.ErrConnClosing) && ctx.Err() == nil {
			select {
			case <-out:
			default:
			}

			out <- runtime.NewErrorMessage(err)
		}
	}()
	return out
}
