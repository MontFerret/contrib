package core

import (
	"errors"
	"fmt"
)

var (
	errWorkbookClosed = errors.New("workbook is closed")
	errSheetDeleted   = errors.New("worksheet has been deleted")
	errSheetStale     = errors.New("worksheet handle is stale")
)

func OperationError(operation string, err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("DOCUMENT::XLSX %s failed: %w", operation, err)
}

func OperationErrorf(operation, format string, args ...any) error {
	return OperationError(operation, fmt.Errorf(format, args...))
}
