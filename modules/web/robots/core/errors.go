package core

import "fmt"

type Stage string

const StageParse Stage = "parse"

// Error reports a robots-specific failure with stage context.
type Error struct {
	Err   error
	Stage Stage
	Msg   string
}

func newError(stage Stage, msg string) error {
	return &Error{
		Stage: stage,
		Msg:   msg,
	}
}

func newErrorf(stage Stage, format string, args ...any) error {
	return newError(stage, fmt.Sprintf(format, args...))
}

func wrapError(stage Stage, err error, msg string) error {
	if err == nil {
		return nil
	}

	return &Error{
		Stage: stage,
		Msg:   msg,
		Err:   err,
	}
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("web::robots: %s: %s: %v", e.Stage, e.Msg, e.Err)
	}

	return fmt.Sprintf("web::robots: %s: %s", e.Stage, e.Msg)
}

func (e *Error) Unwrap() error {
	return e.Err
}
