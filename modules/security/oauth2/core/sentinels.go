package core

import "errors"

var (
	// ErrInvalidExecutor indicates invalid token executor configuration.
	ErrInvalidExecutor = errors.New("invalid OAuth2 token executor")
	// ErrInvalidGrant indicates invalid grant input.
	ErrInvalidGrant = errors.New("invalid OAuth2 grant")
	// ErrInvalidTokenResponse indicates a malformed token endpoint response.
	ErrInvalidTokenResponse = errors.New("invalid OAuth2 token response")
)
