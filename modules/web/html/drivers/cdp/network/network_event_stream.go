package network

import (
	"context"
	"sync"

	"github.com/rs/zerolog"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type networkEventStream struct {
	logger    zerolog.Logger
	observer  *networkObserver
	done      chan struct{}
	eventName string
	options   networkEventOptions
	closeOnce sync.Once
}

func newNetworkEventStream(
	observer *networkObserver,
	logger zerolog.Logger,
	eventName string,
	options networkEventOptions,
) runtime.Stream {
	return &networkEventStream{
		observer:  observer,
		logger:    logger,
		eventName: eventName,
		options:   options,
		done:      make(chan struct{}),
	}
}

func (s *networkEventStream) Close() error {
	s.closeOnce.Do(func() {
		close(s.done)
	})

	return nil
}

func (s *networkEventStream) Read(ctx context.Context) <-chan runtime.Message {
	out := make(chan runtime.Message)

	go func() {
		defer close(out)

		subscriber := s.observer.subscribe()
		defer s.observer.unsubscribe(subscriber.id)

		for {
			select {
			case <-ctx.Done():
				return
			case <-s.done:
				return
			case event := <-subscriber.ch:
				if event.err != nil {
					if !sendNetworkMessage(ctx, s.done, out, runtime.NewErrorMessage(event.err)) {
						return
					}

					continue
				}

				if event.name != s.eventName {
					continue
				}

				payload := buildNetworkEventPayload(ctx, s.logger, event, s.options)
				if !sendNetworkMessage(ctx, s.done, out, runtime.NewValueMessage(payload)) {
					return
				}
			}
		}
	}()

	return out
}
