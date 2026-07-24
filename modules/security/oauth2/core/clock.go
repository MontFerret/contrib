package core

import "time"

// Clock supplies the current time used to calculate token expiry.
type Clock interface {
	Now() time.Time
}

// ClockFunc adapts a function to Clock.
type ClockFunc func() time.Time

// Now returns the function's current time.
func (f ClockFunc) Now() time.Time {
	return f()
}
