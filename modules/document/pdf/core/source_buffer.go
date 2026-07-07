package core

import (
	"context"
	"errors"
	"fmt"
	"io"

	ferretfs "github.com/MontFerret/ferret/v2/pkg/fs"
)

func bufferSource(ctx context.Context, file ferretfs.ReadableFile, path string, size, limit int64) (_ []byte, outErr error) {
	defer func() {
		if err := file.Close(); err != nil {
			outErr = errors.Join(outErr, fmt.Errorf("close PDF document %q: %w", path, err))
		}
	}()

	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if size > limit {
		return nil, fmt.Errorf("PDF document %q is %d bytes, which exceeds the in-memory buffer limit of %d bytes", path, size, limit)
	}

	data, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil {
		return nil, fmt.Errorf("buffer PDF document %q: %w", path, err)
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("PDF document %q exceeds the in-memory buffer limit of %d bytes", path, limit)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return data, nil
}
