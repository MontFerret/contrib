package core

import "fmt"

// Error reports a CSV-specific error with optional row and column context.
type Error struct {
	Err    error
	Msg    string
	Row    int
	Column int
}

func newError(row int, msg string) *Error {
	return &Error{Row: row, Msg: msg}
}

func newErrorf(row int, format string, args ...any) *Error {
	return newError(row, fmt.Sprintf(format, args...))
}

// Error formats the CSV error message with row and optional column metadata.
func (e *Error) Error() string {
	if e.Column > 0 {
		return fmt.Sprintf("csv: row %d, col %d: %s", e.Row, e.Column, e.Msg)
	}

	return fmt.Sprintf("csv: row %d: %s", e.Row, e.Msg)
}

// Unwrap returns the underlying wrapped error.
func (e *Error) Unwrap() error {
	return e.Err
}
