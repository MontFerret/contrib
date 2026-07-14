package core

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestStableErrorCodes(t *testing.T) {
	codes := []ErrorCode{
		ErrProvider,
		ErrAuth,
		ErrRateLimit,
		ErrTimeout,
		ErrContextLimit,
		ErrUnsupportedProvider,
		ErrUnsupportedOperation,
		ErrInvalidOptions,
		ErrInvalidSchema,
		ErrInvalidStructuredOutput,
		ErrSchemaValidation,
		ErrRefusal,
	}

	seen := make(map[ErrorCode]struct{}, len(codes))
	for _, code := range codes {
		if code == "" {
			t.Fatal("error code must not be empty")
		}
		if _, exists := seen[code]; exists {
			t.Fatalf("duplicate error code %s", code)
		}
		seen[code] = struct{}{}
		requireCode(t, NewError(code, "safe"), code)
	}
}

func TestOperationErrorSanitizesUntypedErrors(t *testing.T) {
	err := OperationError("GENERATE", errors.New("secret-api-key"))
	requireCode(t, err, ErrProvider)

	if strings.Contains(err.Error(), "secret") {
		t.Fatalf("error leaked cause: %s", err)
	}
}

func TestOperationErrorPreservesCancellation(t *testing.T) {
	err := OperationError("GENERATE", errors.Join(errors.New("transport wrapper"), context.Canceled))

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}

	if _, ok := CodeOf(err); ok {
		t.Fatalf("cancellation must remain control flow, got %v", err)
	}
}
