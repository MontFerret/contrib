package core

import (
	"errors"
	"fmt"
)

var (
	errConnectionClosed = errors.New("database connection has been closed")
	errTransactionDone  = errors.New("transaction has already been finished")
	errParentClosed     = errors.New("parent database has been closed")
)

// OperationError wraps an error with the DB::SQLITE operation context.
func OperationError(operation string, err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("DB::SQLITE %s failed: %w", operation, err)
}

// OperationErrorf formats a DB::SQLITE operation error.
func OperationErrorf(operation, format string, args ...any) error {
	return fmt.Errorf("DB::SQLITE %s failed: %s", operation, fmt.Sprintf(format, args...))
}
