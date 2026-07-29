package core

import "fmt"

// EntryPathError reports an archive entry name unsafe for extraction.
type EntryPathError struct {
	Name   string
	Reason string
}

func (e *EntryPathError) Error() string {
	return fmt.Sprintf("unsafe archive entry path %q: %s", e.Name, e.Reason)
}
