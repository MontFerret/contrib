package stream

import (
	"bytes"
	"context"
	"errors"
	"io"
	"math"
	"testing"
)

func TestReadAllLimitedLimits(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		limit   int64
		want    string
		wantErr error
	}{
		{
			name:    "negative",
			content: "data",
			limit:   -1,
			wantErr: ErrInvalidLimit,
		},
		{
			name:  "zero empty",
			limit: 0,
		},
		{
			name:    "zero non-empty",
			content: "x",
			limit:   0,
			wantErr: ErrLimitExceeded,
		},
		{
			name:    "below limit",
			content: "data",
			limit:   5,
			want:    "data",
		},
		{
			name:    "exact limit",
			content: "data",
			limit:   4,
			want:    "data",
		},
		{
			name:    "over limit",
			content: "data",
			limit:   3,
			wantErr: ErrLimitExceeded,
		},
		{
			name:    "maximum limit",
			content: "data",
			limit:   math.MaxInt64,
			want:    "data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ReadAllLimited(context.Background(), bytes.NewBufferString(tt.content), tt.limit)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr != nil && got != nil {
				t.Fatalf("expected no partial data, got %q", got)
			}
			if string(got) != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestCopyLimitedLimits(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		content     string
		limit       int64
		want        string
		wantWritten int64
		wantErr     error
	}{
		{
			name:    "negative",
			content: "data",
			limit:   -1,
			wantErr: ErrInvalidLimit,
		},
		{
			name:  "zero empty",
			limit: 0,
		},
		{
			name:        "zero non-empty",
			content:     "x",
			limit:       0,
			wantErr:     ErrLimitExceeded,
			wantWritten: 0,
		},
		{
			name:        "below limit",
			content:     "data",
			limit:       5,
			want:        "data",
			wantWritten: 4,
		},
		{
			name:        "exact limit",
			content:     "data",
			limit:       4,
			want:        "data",
			wantWritten: 4,
		},
		{
			name:        "over limit",
			content:     "data",
			limit:       3,
			want:        "dat",
			wantWritten: 3,
			wantErr:     ErrLimitExceeded,
		},
		{
			name:        "maximum limit",
			content:     "data",
			limit:       math.MaxInt64,
			want:        "data",
			wantWritten: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var destination bytes.Buffer

			written, err := CopyLimited(
				context.Background(),
				&destination,
				bytes.NewBufferString(tt.content),
				tt.limit,
			)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
			if written != tt.wantWritten {
				t.Fatalf("expected %d bytes written, got %d", tt.wantWritten, written)
			}
			if destination.String() != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, destination.String())
			}
		})
	}
}

func TestCopyLimitedConsumesOnlyOneOverflowByte(t *testing.T) {
	t.Parallel()

	source := bytes.NewReader([]byte("abcdef"))
	var destination bytes.Buffer

	written, err := CopyLimited(context.Background(), &destination, source, 4)
	if !errors.Is(err, ErrLimitExceeded) {
		t.Fatalf("expected limit error, got %v", err)
	}
	if written != 4 {
		t.Fatalf("expected 4 bytes written, got %d", written)
	}
	if consumed := source.Size() - int64(source.Len()); consumed != 5 {
		t.Fatalf("expected 5 bytes consumed, got %d", consumed)
	}
}

func TestReadAllLimitedReturnsNoPartialDataOnReadFailure(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("read failed")
	called := false
	source := readerFunc(func(buffer []byte) (int, error) {
		if called {
			return 0, wantErr
		}

		called = true

		return copy(buffer, "data"), nil
	})

	got, err := ReadAllLimited(context.Background(), source, 10)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected read error, got %v", err)
	}
	if got != nil {
		t.Fatalf("expected no partial data, got %q", got)
	}
}

func TestCopyLimitedPreservesReadFailure(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("read failed")
	source := readerFunc(func(buffer []byte) (int, error) {
		return copy(buffer, "x"), wantErr
	})
	var destination bytes.Buffer

	written, err := CopyLimited(context.Background(), &destination, source, 10)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected read error, got %v", err)
	}
	if written != 1 {
		t.Fatalf("expected 1 byte written, got %d", written)
	}
	if destination.String() != "x" {
		t.Fatalf("expected partial write, got %q", destination.String())
	}
}

func TestCopyLimitedPreservesWriteFailure(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("write failed")
	destination := writerFunc(func(buffer []byte) (int, error) {
		return 1, wantErr
	})

	written, err := CopyLimited(context.Background(), destination, bytes.NewBufferString("data"), 10)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected write error, got %v", err)
	}
	if written != 1 {
		t.Fatalf("expected 1 byte written, got %d", written)
	}
}

func TestCopyLimitedReportsShortWrite(t *testing.T) {
	t.Parallel()

	destination := writerFunc(func(buffer []byte) (int, error) {
		return len(buffer) - 1, nil
	})

	written, err := CopyLimited(context.Background(), destination, bytes.NewBufferString("data"), 10)
	if !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("expected short-write error, got %v", err)
	}
	if written != 3 {
		t.Fatalf("expected 3 bytes written, got %d", written)
	}
}

func TestCopyLimitedReportsNoProgress(t *testing.T) {
	t.Parallel()

	t.Run("while copying", func(t *testing.T) {
		t.Parallel()

		source := readerFunc(func([]byte) (int, error) {
			return 0, nil
		})

		_, err := CopyLimited(context.Background(), io.Discard, source, 1)
		if !errors.Is(err, io.ErrNoProgress) {
			t.Fatalf("expected no-progress error, got %v", err)
		}
	})

	t.Run("while probing overflow", func(t *testing.T) {
		t.Parallel()

		calls := 0
		source := readerFunc(func(buffer []byte) (int, error) {
			calls++
			if calls == 1 {
				return copy(buffer, "x"), nil
			}

			return 0, nil
		})

		var destination bytes.Buffer

		written, err := CopyLimited(context.Background(), &destination, source, 1)
		if !errors.Is(err, io.ErrNoProgress) {
			t.Fatalf("expected no-progress error, got %v", err)
		}
		if written != 1 {
			t.Fatalf("expected 1 byte written, got %d", written)
		}
	})
}

func TestCopyLimitedHonorsCancellationBeforeRead(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	read := false
	source := readerFunc(func([]byte) (int, error) {
		read = true

		return 0, io.EOF
	})

	_, err := CopyLimited(ctx, io.Discard, source, 1)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
	if read {
		t.Fatal("source was read after cancellation")
	}
}

func TestCopyLimitedHonorsCancellationBeforeWrite(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	source := readerFunc(func(buffer []byte) (int, error) {
		cancel()

		return copy(buffer, "data"), nil
	})
	wrote := false
	destination := writerFunc(func(buffer []byte) (int, error) {
		wrote = true

		return len(buffer), nil
	})

	written, err := CopyLimited(ctx, destination, source, 10)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
	if written != 0 {
		t.Fatalf("expected no bytes written, got %d", written)
	}
	if wrote {
		t.Fatal("destination was written after cancellation")
	}
}

func TestCopyLimitedHonorsCancellationAtCompletion(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	source := readerFunc(func(buffer []byte) (int, error) {
		return copy(buffer, "x"), io.EOF
	})
	destination := writerFunc(func(buffer []byte) (int, error) {
		cancel()

		return len(buffer), nil
	})

	written, err := CopyLimited(ctx, destination, source, 10)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
	if written != 1 {
		t.Fatalf("expected 1 byte written, got %d", written)
	}
}

func TestCopyLimitedDoesNotCloseStreams(t *testing.T) {
	t.Parallel()

	source := &closeTrackingBuffer{}
	source.WriteString("data")
	destination := &closeTrackingBuffer{}

	if _, err := CopyLimited(context.Background(), destination, source, 10); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if source.closed {
		t.Fatal("source was closed")
	}
	if destination.closed {
		t.Fatal("destination was closed")
	}
}
