package stream

import (
	"errors"
	"io"
	"sync"
)

type readerAt struct {
	source io.ReadSeeker
	mu     sync.Mutex
}

// NewReaderAt adapts source to an io.ReaderAt by serializing seek and read
// operations. Callers must not access source directly while using the adapter.
func NewReaderAt(source io.ReadSeeker) io.ReaderAt {
	return &readerAt{source: source}
}

func (r *readerAt) ReadAt(buffer []byte, offset int64) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, err := r.source.Seek(offset, io.SeekStart); err != nil {
		return 0, err
	}

	n, err := io.ReadFull(r.source, buffer)
	if errors.Is(err, io.ErrUnexpectedEOF) {
		err = io.EOF
	}

	return n, err
}
