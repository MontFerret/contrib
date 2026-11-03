package network

import (
	"context"
	"sync"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type sessionStream struct {
	stream    runtime.Stream
	cancel    context.CancelFunc
	closeErr  error
	closeOnce sync.Once
}

func newSessionStream(stream runtime.Stream, cancel context.CancelFunc) *sessionStream {
	return &sessionStream{stream: stream, cancel: cancel}
}

func (s *sessionStream) Close() error {
	s.closeOnce.Do(func() {
		s.cancel()
		s.closeErr = s.stream.Close()
	})

	return s.closeErr
}
