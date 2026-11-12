package network

import (
	"context"
	"sync"

	"github.com/mafredri/cdp/protocol/target"

	cdpsession "github.com/MontFerret/contrib/modules/web/html/drivers/cdp/session"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type (
	sessionStreamOpener func(context.Context, *cdpsession.Client) (runtime.Stream, error)

	sessionStreamManager interface {
		PageContext() context.Context
		FrameError(target.ID) error
		Snapshot() []*cdpsession.Client
		AddListener(cdpsession.Listener) cdpsession.ListenerID
		RemoveListener(cdpsession.ListenerID)
	}
)

type sessionRuntimeStream struct {
	sessions        sessionStreamManager
	frameID         target.ID
	open            sessionStreamOpener
	ctx             context.Context
	cancel          context.CancelFunc
	streams         map[string]*sessionStream
	initialDetached map[string]struct{}
	messages        chan runtime.Message
	done            chan struct{}
	stopRead        func() bool
	stopSessions    func() bool
	listenerID      cdpsession.ListenerID
	wg              sync.WaitGroup
	mu              sync.Mutex
	readOnce        sync.Once
	closed          bool
	initializing    bool
}

func newSessionRuntimeStream(
	ctx context.Context,
	sessions sessionStreamManager,
	frameID target.ID,
	open sessionStreamOpener,
) (runtime.Stream, error) {
	streamCtx, end := context.WithCancelCause(ctx)
	cancel := func() { end(context.Canceled) }
	s := &sessionRuntimeStream{
		sessions:     sessions,
		frameID:      frameID,
		initializing: true,
		open:         open,
		ctx:          streamCtx,
		cancel:       cancel,
		streams:      make(map[string]*sessionStream),
		messages:     make(chan runtime.Message, 1),
		done:         make(chan struct{}),
	}

	if sessions == nil {
		s.closed = true
		cancel()
		close(s.messages)
		close(s.done)

		return s, nil
	}

	// Register before taking the snapshot so a concurrent attachment cannot be missed.
	if err := ctx.Err(); err != nil {
		cancel()

		return nil, err
	}

	if err := context.Cause(sessions.PageContext()); err != nil {
		cancel()

		return nil, err
	}

	pageCtx := sessions.PageContext()
	s.stopSessions = context.AfterFunc(pageCtx, func() { end(context.Cause(pageCtx)) })
	s.listenerID = sessions.AddListener(s.onSessionEvent)
	go s.cleanup()
	if frameID != "" {
		if err := sessions.FrameError(frameID); err != nil {
			_ = s.Close()

			return nil, err
		}
	}

	for _, client := range sessions.Snapshot() {
		if err := s.attach(client); err != nil {
			_ = s.Close()

			return nil, err
		}
	}

	s.mu.Lock()
	s.initializing = false
	s.initialDetached = nil
	s.mu.Unlock()

	return s, nil
}

func (s *sessionRuntimeStream) Close() error {
	s.cancel()
	<-s.done

	return nil
}

func (s *sessionRuntimeStream) Read(ctx context.Context) <-chan runtime.Message {
	s.readOnce.Do(func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		if !s.closed {
			s.stopRead = context.AfterFunc(ctx, s.cancel)
		}
	})

	return s.messages
}

func (s *sessionRuntimeStream) onSessionEvent(event cdpsession.Event) {
	switch event.Kind {
	case cdpsession.EventFailed:
		if s.frameID != "" && event.Client != nil && !event.Client.BelongsToPopup() && event.Client.TargetID == s.frameID {
			s.reportError(event.Err)
		}
	case cdpsession.EventAttached:
		if err := s.attach(event.Client); err != nil {
			s.reportError(err)
		}
	case cdpsession.EventDetached:
		s.detach(event.Client)
	}
}

func (s *sessionRuntimeStream) attach(client *cdpsession.Client) error {
	if client == nil || client.BelongsToPopup() || !client.SupportsNetwork() {
		return nil
	}

	key := string(client.ID)
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := context.Cause(s.ctx); err != nil {
		return err
	}

	if s.closed || s.streams[key] != nil {
		return nil
	}

	// The initial snapshot can outlive a session's detach notification.
	if _, detached := s.initialDetached[key]; detached {
		return nil
	}

	ctx, cancel := context.WithCancel(s.ctx)
	stream, err := s.open(ctx, client)
	if err != nil {
		cause := context.Cause(ctx)
		cancel()

		if cause != nil {
			return cause
		}

		// A snapshot can contain a child whose route closed during subscription.
		if client.Conn != nil && client.Conn.Context().Err() != nil {
			return nil
		}

		return err
	}

	if stream == nil {
		cancel()

		return nil
	}

	entry := newSessionStream(stream, cancel)
	s.streams[key] = entry
	// Add while holding mu: cleanup prevents further additions before waiting.
	s.wg.Add(1)
	go s.consume(ctx, key, entry)

	return nil
}

func (s *sessionRuntimeStream) consume(ctx context.Context, key string, entry *sessionStream) {
	defer s.wg.Done()
	defer s.removeStream(key, entry)
	defer entry.Close()

	for message := range entry.stream.Read(ctx) {
		select {
		case <-ctx.Done():
			return
		case s.messages <- message:
		}
	}
}

func (s *sessionRuntimeStream) reportError(err error) {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()

		return
	}

	s.wg.Add(1)
	s.mu.Unlock()

	// Session callbacks must not wait for a caller to begin reading.
	go func() {
		defer s.wg.Done()

		select {
		case <-s.ctx.Done():
		case s.messages <- runtime.NewErrorMessage(err):
		}
	}()
}

func (s *sessionRuntimeStream) detach(client *cdpsession.Client) {
	if client == nil || client.BelongsToPopup() || !client.SupportsNetwork() {
		return
	}

	key := string(client.ID)
	s.mu.Lock()
	if s.initializing {
		if s.initialDetached == nil {
			s.initialDetached = make(map[string]struct{})
		}

		s.initialDetached[key] = struct{}{}
	}

	entry := s.streams[key]
	delete(s.streams, key)
	s.mu.Unlock()

	if entry != nil {
		_ = entry.Close()
	}
}

func (s *sessionRuntimeStream) removeStream(key string, entry *sessionStream) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.streams[key] == entry {
		delete(s.streams, key)
	}
}

func (s *sessionRuntimeStream) cleanup() {
	<-s.ctx.Done()
	s.stopSessions()
	s.sessions.RemoveListener(s.listenerID)

	s.mu.Lock()
	s.closed = true
	if s.stopRead != nil {
		s.stopRead()
	}

	entries := s.streams
	s.streams = nil
	s.mu.Unlock()

	for _, entry := range entries {
		_ = entry.Close()
	}

	s.wg.Wait()
	closePageMessages(s.sessions.PageContext(), s.messages)
	close(s.done)
}
