package core

import "fmt"

// Error reports a TOML-specific decode or encode failure.
type Error struct {
	Err error
	Msg string
}

func newError(msg string) error {
	return &Error{Msg: msg}
}

func newErrorf(format string, args ...any) error {
	return newError(fmt.Sprintf(format, args...))
}

func wrapError(err error, msg string) error {
	if err == nil {
		return nil
	}

	return &Error{
		Err: err,
		Msg: msg,
	}
}

// Error formats the TOML error with the module prefix.
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("toml: %s: %v", e.Msg, e.Err)
	}

	return fmt.Sprintf("toml: %s", e.Msg)
}

// Unwrap returns the wrapped error, when present.
func (e *Error) Unwrap() error {
	return e.Err
}
