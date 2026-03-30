package core

import "fmt"

// Error reports XML-specific decode and encode failures.
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

// Error formats the XML error with the module prefix.
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("xml: %s: %v", e.Msg, e.Err)
	}

	return fmt.Sprintf("xml: %s", e.Msg)
}

// Unwrap returns the wrapped error, when present.
func (e *Error) Unwrap() error {
	return e.Err
}
