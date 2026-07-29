package core

import (
	"context"
	"errors"
	"fmt"
	"io"

	commonstream "github.com/MontFerret/contrib/pkg/common/stream"
)

func readBounded(ctx context.Context, reader io.Reader, name string, limit int64) ([]byte, error) {
	data, err := commonstream.ReadAllLimited(ctx, reader, limit)
	if err != nil {
		if errors.Is(err, commonstream.ErrLimitExceeded) {
			return nil, &LimitError{Entry: name, Size: -1, Limit: limit}
		}

		return nil, fmt.Errorf("read archive entry %q: %w", name, err)
	}

	return data, nil
}

func copyBounded(ctx context.Context, writer io.Writer, reader io.Reader, name string, limit int64) error {
	_, err := commonstream.CopyLimited(ctx, writer, reader, limit)
	if err == nil {
		return nil
	}

	if errors.Is(err, commonstream.ErrLimitExceeded) {
		return &LimitError{Entry: name, Size: -1, Limit: limit}
	}

	return fmt.Errorf("copy archive entry %q: %w", name, err)
}
