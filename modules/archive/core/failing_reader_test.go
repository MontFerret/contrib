package core

import "errors"

var errSyntheticRead = errors.New("synthetic read failure")

type failingReader struct {
	data []byte
	read bool
}

func (r *failingReader) Read(p []byte) (int, error) {
	if r.read {
		return 0, errSyntheticRead
	}
	r.read = true

	return copy(p, r.data), nil
}
