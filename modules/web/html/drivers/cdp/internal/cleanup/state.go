// Package cleanup coordinates one teardown operation with independent waiters.
package cleanup

import (
	"context"
	"sync"
)

// State starts one cleanup operation. Each caller can stop waiting separately
// while the owner's cleanup continues toward its retained result.
type State struct {
	once sync.Once
	done chan struct{}
	err  error
}

func (s *State) Run(ctx context.Context, closeOwner func(context.Context) error) error {
	s.once.Do(func() {
		s.done = make(chan struct{})
		go func() {
			s.err = closeOwner(ctx)
			close(s.done)
		}()
	})

	select {
	case <-s.done:
		return s.err
	default:
	}

	select {
	case <-s.done:
		return s.err
	case <-ctx.Done():
		return ctx.Err()
	}
}
