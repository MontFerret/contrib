package core

import "fmt"

// TOMLError reports a TOML-specific decode or encode failure.
type TOMLError struct {
	Err error
	Msg string
}

// Error formats the TOML error with the module prefix.
func (e *TOMLError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("toml: %s: %v", e.Msg, e.Err)
	}

	return fmt.Sprintf("toml: %s", e.Msg)
}

// Unwrap returns the wrapped error, when present.
func (e *TOMLError) Unwrap() error {
	return e.Err
}

func newTOMLError(msg string) error {
	return &TOMLError{Msg: msg}
}

func newTOMLErrorf(format string, args ...any) error {
	return newTOMLError(fmt.Sprintf(format, args...))
}

func wrapTOMLError(err error, msg string) error {
	if err == nil {
		return nil
	}

	return &TOMLError{
		Err: err,
		Msg: msg,
	}
}
