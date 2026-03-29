package types

import "fmt"

// CSVError reports a CSV-specific error with optional row and column context.
type CSVError struct {
	Err    error
	Msg    string
	Row    int
	Column int
}

// Error formats the CSV error message with row and optional column metadata.
func (e *CSVError) Error() string {
	if e.Column > 0 {
		return fmt.Sprintf("csv: row %d, col %d: %s", e.Row, e.Column, e.Msg)
	}

	return fmt.Sprintf("csv: row %d: %s", e.Row, e.Msg)
}

// Unwrap returns the underlying wrapped error.
func (e *CSVError) Unwrap() error {
	return e.Err
}

func newCSVError(row int, msg string) *CSVError {
	return &CSVError{Row: row, Msg: msg}
}

func newCSVErrorf(row int, format string, args ...any) *CSVError {
	return newCSVError(row, fmt.Sprintf(format, args...))
}
