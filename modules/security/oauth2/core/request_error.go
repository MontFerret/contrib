package core

// tokenRequestError preserves an underlying typed error while keeping its
// potentially untrusted message secret-free.
type tokenRequestError struct {
	cause   error
	message string
}

func newTokenRequestError(cause error, secrets []string) *tokenRequestError {
	return &tokenRequestError{
		cause:   cause,
		message: "oauth2 token request: " + sanitizeTokenText(cause.Error(), secrets),
	}
}

func (e *tokenRequestError) Error() string {
	return e.message
}

func (e *tokenRequestError) Unwrap() error {
	return e.cause
}
