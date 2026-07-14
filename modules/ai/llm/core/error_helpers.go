package core

import (
	"context"
	"errors"
)

// CodeOf returns the stable AI::LLM error code, when present.
func CodeOf(err error) (ErrorCode, bool) {
	var typed *Error
	if !errors.As(err, &typed) {
		return "", false
	}

	return typed.Code, true
}

// OperationError annotates a typed error without exposing an underlying provider error.
// Cancellation remains untyped control flow so callers can detect it with errors.Is.
func OperationError(operation string, err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, context.Canceled) {
		return context.Canceled
	}

	if typed, ok := errors.AsType[*Error](err); ok {
		return typed.WithOperation(operation)
	}

	return NewError(ErrProvider, "provider request failed").WithOperation(operation)
}
