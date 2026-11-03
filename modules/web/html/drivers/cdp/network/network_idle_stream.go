package network

import (
	"context"
	"time"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type networkIdleStream struct {
	observer     *networkObserver
	subscription *networkSubscription
	options      networkIdleOptions
}

func newNetworkIdleStream(ctx context.Context, observer *networkObserver, options networkIdleOptions) runtime.Stream {
	return &networkIdleStream{
		observer:     observer,
		options:      options,
		subscription: newNetworkSubscription(ctx, observer),
	}
}

func (s *networkIdleStream) Close() error {
	return s.subscription.Close()
}

func (s *networkIdleStream) Read(ctx context.Context) <-chan runtime.Message {
	out := make(chan runtime.Message)

	go func() {
		defer close(out)

		subscriber := s.subscription.subscriber
		defer s.Close()

		active := s.observer.snapshotActive(s.options.types)
		timer := s.newTimer(len(active))
		defer stopIdleTimer(timer)

		for {
			var timerCh <-chan time.Time
			if timer != nil {
				timerCh = timer.C
			}

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

				s.updateActive(active, event)
				timer = s.resetTimer(timer, len(active))
			case <-timerCh:
				timer = nil

				if !sendNetworkMessage(
					ctx,
					subscriber.done,
					out,
					runtime.NewValueMessage(buildNetworkIdlePayload(s.options, len(active))),
				) {
					return
				}
			}
		}
	}()

	return out
}

func (s *networkIdleStream) newTimer(inflight int) *time.Timer {
	if inflight > s.options.maxInflight {
		return nil
	}

	return time.NewTimer(s.options.quiet)
}

func (s *networkIdleStream) resetTimer(timer *time.Timer, inflight int) *time.Timer {
	if inflight > s.options.maxInflight {
		stopIdleTimer(timer)
		return nil
	}

	if timer == nil {
		return time.NewTimer(s.options.quiet)
	}

	stopIdleTimer(timer)
	timer.Reset(s.options.quiet)

	return timer
}

func (s *networkIdleStream) updateActive(active map[string]networkEvent, event networkEvent) {
	switch event.name {
	case drivers.NetworkRequestStartedEvent:
		if !s.matchesType(event.resourceType) {
			return
		}

		active[networkRequestKey(event.sessionKey, event.requestID)] = event
	case drivers.NetworkRequestFinishedEvent, drivers.NetworkRequestFailedEvent:
		delete(active, networkRequestKey(event.sessionKey, event.requestID))
	case networkSessionDetachedEvent:
		for key, activeEvent := range active {
			if activeEvent.sessionKey == event.sessionKey {
				delete(active, key)
			}
		}
	}
}

func (s *networkIdleStream) matchesType(resourceType string) bool {
	if len(s.options.types) == 0 {
		return true
	}

	_, exists := s.options.types[resourceType]

	return exists
}
