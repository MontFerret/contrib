package core

import (
	"errors"
	"sync"
)

// SessionScope owns every local session created during one Ferret execution.
type SessionScope struct {
	sessions map[uint64]Session
	mu       sync.Mutex
	closed   bool
}

// NewSessionScope creates an empty per-run session scope.
func NewSessionScope() *SessionScope {
	return &SessionScope{sessions: make(map[uint64]Session)}
}

// Track adds a session to this execution scope.
func (s *SessionScope) Track(session Session) error {
	if session == nil {
		return NewError(ErrInvalidOptions, "session must not be nil")
	}

	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		_ = session.Close()

		return NewError(ErrProvider, "session scope is closed")
	}

	s.sessions[session.ResourceID()] = session
	s.mu.Unlock()

	return nil
}

// Len returns the number of sessions currently owned by the scope.
func (s *SessionScope) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.sessions)
}

// Close closes all tracked sessions and clears the scope.
func (s *SessionScope) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()

		return nil
	}

	s.closed = true
	sessions := make([]Session, 0, len(s.sessions))

	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}

	clear(s.sessions)
	s.mu.Unlock()

	var errs []error
	for _, session := range sessions {
		if err := session.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
