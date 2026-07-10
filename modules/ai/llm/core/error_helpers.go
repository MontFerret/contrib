package core

import "errors"

// CodeOf returns the stable AI::LLM error code, when present.
func CodeOf(err error) (ErrorCode, bool) {
	var typed *Error
	if !errors.As(err, &typed) {
		return "", false
	}

	return typed.Code, true
}

// OperationError annotates a typed error without exposing an underlying provider error.
func OperationError(operation string, err error) error {
	if err == nil {
		return nil
	}

	var typed *Error
	if errors.As(err, &typed) {
		return typed.WithOperation(operation)
	}

	return NewError(ErrProvider, "provider request failed").WithOperation(operation)
}
