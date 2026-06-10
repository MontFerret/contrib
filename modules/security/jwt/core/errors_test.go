package core

import (
	"errors"
	"testing"
)

func TestError(t *testing.T) {
	err := newError(ErrInvalidToken, "test message")
	if err.Error() != "jwt.invalid_token: test message" {
		t.Errorf("Error() = %v", err.Error())
	}

	wrapped := wrapError(ErrInvalidToken, "wrapped", errors.New("inner"))
	if wrapped.Error() != "jwt.invalid_token: wrapped" {
		t.Errorf("wrapError Error() = %v", wrapped.Error())
	}

	if errors.Unwrap(wrapped).Error() != "inner" {
		t.Errorf("Unwrap() = %v", errors.Unwrap(wrapped).Error())
	}
}

func TestOperationError(t *testing.T) {
	err := errors.New("fail")
	opErr := OperationError("SIGN", err)
	if opErr.Error() != "SECURITY::JWT SIGN failed: fail" {
		t.Errorf("OperationError() = %v", opErr.Error())
	}

	if OperationError("SIGN", nil) != nil {
		t.Error("OperationError(nil) should be nil")
	}
}
