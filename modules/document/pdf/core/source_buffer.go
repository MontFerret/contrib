package core

import (
	"context"
	"errors"
	"fmt"

	commonstream "github.com/MontFerret/contrib/pkg/common/stream"
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

	data, err := commonstream.ReadAllLimited(ctx, file, limit)
	if err != nil {
		if errors.Is(err, commonstream.ErrLimitExceeded) {
			return nil, fmt.Errorf("PDF document %q exceeds the in-memory buffer limit of %d bytes", path, limit)
		}
		if contextErr := ctx.Err(); contextErr != nil {
			return nil, contextErr
		}

		return nil, fmt.Errorf("buffer PDF document %q: %w", path, err)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return data, nil
}
