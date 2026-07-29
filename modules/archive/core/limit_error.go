package core

import "fmt"

// LimitError reports a declared or observed archive size over a configured cap.
type LimitError struct {
	Entry string
	Size  int64
	Limit int64
}

func (e *LimitError) Error() string {
	if e.Size >= 0 {
		return fmt.Sprintf("archive entry %q is %d bytes, exceeding the limit of %d bytes", e.Entry, e.Size, e.Limit)
	}

	return fmt.Sprintf("archive entry %q exceeds the limit of %d bytes", e.Entry, e.Limit)
}
