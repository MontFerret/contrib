package core

import (
	"errors"
	"fmt"
)

var errConnectionClosed = errors.New("redis connection has been closed")

// OperationError wraps an error with the DB::REDIS operation context.
func OperationError(operation string, err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("DB::REDIS %s failed: %w", operation, err)
}

// OperationErrorf formats a DB::REDIS operation error.
func OperationErrorf(operation, format string, args ...any) error {
	return fmt.Errorf("DB::REDIS %s failed: %s", operation, fmt.Sprintf(format, args...))
}
