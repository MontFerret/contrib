package core

import "fmt"

// YAMLError reports a YAML-specific decode or encode failure.
type YAMLError struct {
	Err error
	Msg string
}

// Error formats the YAML error with the module prefix.
func (e *YAMLError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("yaml: %s: %v", e.Msg, e.Err)
	}

	return fmt.Sprintf("yaml: %s", e.Msg)
}

// Unwrap returns the wrapped error, when present.
func (e *YAMLError) Unwrap() error {
	return e.Err
}

func newYAMLError(msg string) error {
	return &YAMLError{Msg: msg}
}

func newYAMLErrorf(format string, args ...any) error {
	return newYAMLError(fmt.Sprintf(format, args...))
}

func wrapYAMLError(err error, msg string) error {
	if err == nil {
		return nil
	}

	return &YAMLError{
		Err: err,
		Msg: msg,
	}
}
