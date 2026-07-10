package core

import "context"

type sessionScopeContextKey struct{}

// WithSessionScope installs a fresh local-session owner into an execution context.
func WithSessionScope(ctx context.Context) context.Context {
	return context.WithValue(ctx, sessionScopeContextKey{}, NewSessionScope())
}

// SessionScopeFrom returns the per-run session scope, when installed.
func SessionScopeFrom(ctx context.Context) (*SessionScope, bool) {
	scope, ok := ctx.Value(sessionScopeContextKey{}).(*SessionScope)

	return scope, ok
}

// TrackSession registers a newly created local session with the current execution.
func TrackSession(ctx context.Context, session Session) error {
	if scope, ok := SessionScopeFrom(ctx); ok {
		return scope.Track(session)
	}

	return nil
}

// CloseSessionScope closes and clears all sessions created during an execution.
func CloseSessionScope(ctx context.Context) error {
	if scope, ok := SessionScopeFrom(ctx); ok {
		return scope.Close()
	}

	return nil
}
