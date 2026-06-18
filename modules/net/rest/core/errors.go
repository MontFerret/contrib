package core

import "fmt"

// OperationError wraps an error with the HTTP operation context.
func OperationError(operation string, err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("HTTP %s failed: %w", operation, err)
}

// OperationErrorf formats an HTTP operation error.
func OperationErrorf(operation, format string, args ...any) error {
	return fmt.Errorf("HTTP %s failed: %s", operation, fmt.Sprintf(format, args...))
}
