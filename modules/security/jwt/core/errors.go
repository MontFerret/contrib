package core

import (
	"fmt"
)

const (
	ErrInvalidToken         = "jwt.invalid_token"
	ErrInvalidSignature     = "jwt.invalid_signature"
	ErrUnexpectedAlgorithm  = "jwt.unexpected_algorithm"
	ErrUnsupportedAlgorithm = "jwt.unsupported_algorithm"
	ErrExpired              = "jwt.expired"
	ErrNotYetValid          = "jwt.not_yet_valid"
	ErrIssuerMismatch       = "jwt.issuer_mismatch"
	ErrAudienceMismatch     = "jwt.audience_mismatch"
	ErrSubjectMismatch      = "jwt.subject_mismatch"
	ErrClaimMissing         = "jwt.claim_missing"
	ErrInvalidKey           = "jwt.invalid_key"
)

// Error reports a JWT-specific failure with a stable code and safe message.
type Error struct {
	Code string
	Msg  string
	Err  error
}

func newError(code, msg string) error {
	return &Error{Code: code, Msg: msg}
}

func wrapError(code, msg string, err error) error {
	return &Error{Code: code, Msg: msg, Err: err}
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s", e.Code, e.Msg)
	}

	return fmt.Sprintf("%s: %s", e.Code, e.Msg)
}

func (e *Error) Unwrap() error {
	return e.Err
}

// OperationError wraps an error with SECURITY::JWT operation context.
func OperationError(operation string, err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("SECURITY::JWT %s failed: %w", operation, err)
}

// NewInvalidKeyError creates a new error indicating an invalid key with a specific error message.
func NewInvalidKeyError(msg string) error {
	return wrapError(ErrInvalidKey, msg, nil)
}
