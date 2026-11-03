package network

import (
	"context"
	"sync"
)

type networkSubscription struct {
	subscriber *networkEventSubscriber
	observer   *networkObserver
	stop       func() bool
	mu         sync.Mutex
}

func newNetworkSubscription(ctx context.Context, observer *networkObserver) *networkSubscription {
	s := &networkSubscription{
		subscriber: observer.subscribe(),
		observer:   observer,
	}

	// An already-canceled context can run the callback before AfterFunc returns.
	s.mu.Lock()
	s.stop = context.AfterFunc(ctx, func() { _ = s.Close() })
	s.mu.Unlock()

	return s
}

func (s *networkSubscription) Close() error {
	s.mu.Lock()
	stop := s.stop
	s.stop = nil
	s.mu.Unlock()

	if stop != nil {
		stop()
	}

	s.observer.unsubscribe(s.subscriber.id)

	return nil
}
