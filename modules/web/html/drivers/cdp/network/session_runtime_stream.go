package network

import (
	"context"
	"sync"

	"github.com/rs/zerolog"

	cdpsession "github.com/MontFerret/contrib/modules/web/html/drivers/cdp/session"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type sessionStreamOpener func(context.Context, *cdpsession.Client) (runtime.Stream, error)

type sessionRuntimeStream struct {
	logger     zerolog.Logger
	sessions   *cdpsession.Manager
	open       sessionStreamOpener
	streams    map[string]runtime.Stream
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	listenerID cdpsession.ListenerID
	mu         sync.Mutex
}

func newSessionRuntimeStream(
	logger zerolog.Logger,
	sessions *cdpsession.Manager,
	open sessionStreamOpener,
) runtime.Stream {
	return &sessionRuntimeStream{
		logger:   logger,
		sessions: sessions,
		open:     open,
		streams:  make(map[string]runtime.Stream),
	}
}

func (s *sessionRuntimeStream) Close() error {
	s.mu.Lock()
	cancel := s.cancel
	s.cancel = nil
	s.mu.Unlock()

	if cancel != nil {
		cancel()
	}

	return nil
}

func (s *sessionRuntimeStream) Read(ctx context.Context) <-chan runtime.Message {
	out := make(chan runtime.Message)
	if s.sessions == nil {
		close(out)
		return out
	}

	streamCtx, cancel := context.WithCancel(ctx)

	s.mu.Lock()
	s.cancel = cancel
	s.mu.Unlock()

	s.listenerID = s.sessions.AddListener(func(event cdpsession.Event) {
		switch event.Kind {
		case cdpsession.EventAttached:
			s.attach(streamCtx, event.Client, out)
		case cdpsession.EventDetached:
			s.detach(event.Client)
		}
	})

	for _, client := range s.sessions.Snapshot() {
		s.attach(streamCtx, client, out)
	}

	go func() {
		<-streamCtx.Done()
		s.sessions.RemoveListener(s.listenerID)
		s.closeStreams()
		s.wg.Wait()
		close(out)
	}()

	return out
}

func (s *sessionRuntimeStream) attach(ctx context.Context, client *cdpsession.Client, out chan<- runtime.Message) {
	if ctx.Err() != nil || client == nil {
		return
	}

	key := string(client.ID)

	s.mu.Lock()
	if _, exists := s.streams[key]; exists {
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()

	stream, err := s.open(ctx, client)
	if err != nil {
		select {
		case <-ctx.Done():
		case out <- runtime.NewErrorMessage(err):
		}
		return
	}

	s.mu.Lock()
	s.streams[key] = stream
	s.mu.Unlock()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer s.removeStream(key)

		for evt := range stream.Read(ctx) {
			select {
			case <-ctx.Done():
				return
			case out <- evt:
			}
		}
	}()
}

func (s *sessionRuntimeStream) detach(client *cdpsession.Client) {
	if client == nil {
		return
	}

	key := string(client.ID)

	s.mu.Lock()
	stream, exists := s.streams[key]
	if exists {
		delete(s.streams, key)
	}
	s.mu.Unlock()

	if exists {
		_ = stream.Close()
	}
}

func (s *sessionRuntimeStream) closeStreams() {
	s.mu.Lock()
	streams := make([]runtime.Stream, 0, len(s.streams))
	for key, stream := range s.streams {
		streams = append(streams, stream)
		delete(s.streams, key)
	}
	s.mu.Unlock()

	for _, stream := range streams {
		_ = stream.Close()
	}
}

func (s *sessionRuntimeStream) removeStream(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.streams, key)
}
