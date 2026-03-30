package drivers

import "errors"

var (
	ErrDetached = errors.New("element detached")
	ErrNotFound = errors.New("element(s) not found")
)
