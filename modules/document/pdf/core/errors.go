package core

import (
	"errors"
	"fmt"
)

var (
	errDocumentClosed = errors.New("document is closed")
	errInvalidPage    = errors.New("page is not available")
)

// OperationError wraps an implementation error with a DOCUMENT::PDF operation
// label suitable for Ferret callers.
func OperationError(operation string, err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("DOCUMENT::PDF %s failed: %w", operation, err)
}

// OperationErrorf formats and wraps an operation error.
func OperationErrorf(operation, format string, args ...any) error {
	return OperationError(operation, fmt.Errorf(format, args...))
}
