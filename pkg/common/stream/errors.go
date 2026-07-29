package stream

import "errors"

var (
	// ErrLimitExceeded indicates that a stream contains more data than allowed.
	ErrLimitExceeded = errors.New("stream limit exceeded")
	// ErrInvalidLimit indicates that a negative stream limit was supplied.
	ErrInvalidLimit = errors.New("stream limit must not be negative")
)
