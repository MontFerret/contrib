package core

import "fmt"

// XMLError reports XML-specific decode and encode failures.
type XMLError struct {
	Err error
	Msg string
}

// Error formats the XML error with the module prefix.
func (e *XMLError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("xml: %s: %v", e.Msg, e.Err)
	}

	return fmt.Sprintf("xml: %s", e.Msg)
}

// Unwrap returns the wrapped error, when present.
func (e *XMLError) Unwrap() error {
	return e.Err
}

func newXMLError(msg string) error {
	return &XMLError{Msg: msg}
}

func newXMLErrorf(format string, args ...any) error {
	return newXMLError(fmt.Sprintf(format, args...))
}

func wrapXMLError(err error, msg string) error {
	if err == nil {
		return nil
	}

	return &XMLError{
		Err: err,
		Msg: msg,
	}
}
