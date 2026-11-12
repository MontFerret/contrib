package network

import (
	"context"

	"github.com/rs/zerolog"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type networkEventStream struct {
	logger       zerolog.Logger
	subscription *networkSubscription
	eventName    string
	options      networkEventOptions
}

func newNetworkEventStream(
	ctx context.Context,
	observer *networkObserver,
	logger zerolog.Logger,
	eventName string,
	options networkEventOptions,
) runtime.Stream {
	return &networkEventStream{
		logger:       logger,
		eventName:    eventName,
		options:      options,
		subscription: newNetworkSubscription(ctx, observer),
	}
}

func (s *networkEventStream) Close() error {
	return s.subscription.Close()
}

func (s *networkEventStream) Read(ctx context.Context) <-chan runtime.Message {
	out := make(chan runtime.Message, 1)

	go func() {
		defer closePageMessages(s.subscription.observer.sessions.PageContext(), out)

		subscriber := s.subscription.subscriber
		defer s.Close()

		for {
			select {
			case <-ctx.Done():
				return
			case <-subscriber.done:
				return
			case event := <-subscriber.ch:
				if event.err != nil {
					if !sendNetworkMessage(ctx, subscriber.done, out, runtime.NewErrorMessage(event.err)) {
						return
					}

					continue
				}

				if event.name != s.eventName {
					continue
				}

				payload := buildNetworkEventPayload(ctx, s.logger, event, s.options)
				if !sendNetworkMessage(ctx, subscriber.done, out, runtime.NewValueMessage(payload)) {
					return
				}
			}
		}
	}()

	return out
}
