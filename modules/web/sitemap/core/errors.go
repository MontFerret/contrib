package core

import "fmt"

type Stage string

const (
	StageFetch  Stage = "fetch"
	StageParse  Stage = "parse"
	StageExpand Stage = "expand"
)

// Error reports a sitemap-specific failure with URL and stage context.
type Error struct {
	Err   error
	URL   string
	Stage Stage
	Msg   string
}

func newError(url string, stage Stage, msg string) error {
	return &Error{
		URL:   url,
		Stage: stage,
		Msg:   msg,
	}
}

func newErrorf(url string, stage Stage, format string, args ...any) error {
	return newError(url, stage, fmt.Sprintf(format, args...))
}

func wrapError(url string, stage Stage, err error, msg string) error {
	if err == nil {
		return nil
	}

	return &Error{
		URL:   url,
		Stage: stage,
		Msg:   msg,
		Err:   err,
	}
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("web::sitemap: %s %s: %s: %v", e.Stage, e.URL, e.Msg, e.Err)
	}

	return fmt.Sprintf("web::sitemap: %s %s: %s", e.Stage, e.URL, e.Msg)
}

func (e *Error) Unwrap() error {
	return e.Err
}
