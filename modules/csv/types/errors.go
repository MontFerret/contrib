package types

import "fmt"

type CSVError struct {
	Row    int
	Column int
	Msg    string
	Err    error
}

func (e *CSVError) Error() string {
	if e.Column > 0 {
		return fmt.Sprintf("csv: row %d, col %d: %s", e.Row, e.Column, e.Msg)
	}

	return fmt.Sprintf("csv: row %d: %s", e.Row, e.Msg)
}

func (e *CSVError) Unwrap() error {
	return e.Err
}

func newCSVError(row int, msg string) *CSVError {
	return &CSVError{Row: row, Msg: msg}
}

func newCSVErrorf(row int, format string, args ...any) *CSVError {
	return &CSVError{Row: row, Msg: fmt.Sprintf(format, args...)}
}
