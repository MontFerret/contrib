package stream

import (
	"bytes"
	"context"
	"io"
)

const copyBufferSize = 32 * 1024

// ReadAllLimited reads source until EOF while enforcing limit.
//
// It returns no partial data on failure and does not close source.
func ReadAllLimited(ctx context.Context, source io.Reader, limit int64) ([]byte, error) {
	var buffer bytes.Buffer

	if _, err := CopyLimited(ctx, &buffer, source, limit); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// CopyLimited copies source to destination while enforcing limit.
//
// It may write limit bytes before returning ErrLimitExceeded. It does not close
// either stream.
func CopyLimited(ctx context.Context, destination io.Writer, source io.Reader, limit int64) (int64, error) {
	if limit < 0 {
		return 0, ErrInvalidLimit
	}

	buffer := make([]byte, copyBufferSize)
	remaining := limit
	var written int64

	for {
		if err := ctx.Err(); err != nil {
			return written, err
		}

		if remaining == 0 {
			var extra [1]byte

			n, err := source.Read(extra[:])

			if n > 0 {
				return written, ErrLimitExceeded
			}

			if err != nil {
				if err == io.EOF {
					if contextErr := ctx.Err(); contextErr != nil {
						return written, contextErr
					}

					return written, nil
				}

				return written, err
			}

			return written, io.ErrNoProgress
		}

		readSize := int64(len(buffer))

		if remaining < readSize {
			readSize = remaining
		}

		n, readErr := source.Read(buffer[:readSize])
		if n > 0 {
			if err := ctx.Err(); err != nil {
				return written, err
			}

			writeCount, writeErr := destination.Write(buffer[:n])

			if writeCount > 0 {
				written += int64(writeCount)
				remaining -= int64(writeCount)
			}

			if writeErr != nil {
				return written, writeErr
			}

			if writeCount != n {
				return written, io.ErrShortWrite
			}
		}

		if readErr != nil {
			if readErr == io.EOF {
				if err := ctx.Err(); err != nil {
					return written, err
				}

				return written, nil
			}

			return written, readErr
		}

		if n == 0 {
			return written, io.ErrNoProgress
		}
	}
}
