package core

import (
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
