package stream

import (
	"bytes"
	"errors"
	"io"
	"sync"
	"testing"
)

func TestReaderAtReadsOffsets(t *testing.T) {
	t.Parallel()

	reader := NewReaderAt(bytes.NewReader([]byte("0123456789")))
	buffer := make([]byte, 4)

	n, err := reader.ReadAt(buffer, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != len(buffer) {
		t.Fatalf("expected %d bytes, got %d", len(buffer), n)
	}
	if string(buffer) != "3456" {
		t.Fatalf("expected %q, got %q", "3456", buffer)
	}
}

func TestReaderAtReturnsEOFForPartialRead(t *testing.T) {
	t.Parallel()

	reader := NewReaderAt(bytes.NewReader([]byte("data")))
	buffer := make([]byte, 4)

	n, err := reader.ReadAt(buffer, 2)
	if !errors.Is(err, io.EOF) {
		t.Fatalf("expected EOF, got %v", err)
	}
	if errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("expected unexpected EOF to be normalized, got %v", err)
	}
	if n != 2 {
		t.Fatalf("expected 2 bytes, got %d", n)
	}
	if string(buffer[:n]) != "ta" {
		t.Fatalf("expected %q, got %q", "ta", buffer[:n])
	}
}

func TestReaderAtPreservesSeekFailure(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("seek failed")
	reader := NewReaderAt(&failingReadSeeker{err: wantErr})

	n, err := reader.ReadAt(make([]byte, 1), 1)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected seek error, got %v", err)
	}
	if n != 0 {
		t.Fatalf("expected no bytes, got %d", n)
	}
}

func TestReaderAtSerializesConcurrentReads(t *testing.T) {
	t.Parallel()

	const (
		content    = "0123456789abcdefghijklmnopqrstuvwxyz"
		iterations = 100
	)

	reader := NewReaderAt(bytes.NewReader([]byte(content)))
	offsets := []int64{0, 3, 8, 12, 20, 28}
	start := make(chan struct{})
	errs := make(chan error, len(offsets))
	var workers sync.WaitGroup

	for _, offset := range offsets {
		offset := offset
		workers.Add(1)

		go func() {
			defer workers.Done()
			<-start

			want := content[offset : offset+4]
			for range iterations {
				buffer := make([]byte, 4)

				n, err := reader.ReadAt(buffer, offset)
				if err != nil {
					errs <- err

					return
				}
				if n != len(buffer) || string(buffer) != want {
					errs <- errors.New("concurrent read returned incorrect data")

					return
				}
			}
		}()
	}

	close(start)
	workers.Wait()
	close(errs)

	for err := range errs {
		t.Fatal(err)
	}
}
