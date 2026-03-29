package types

import (
	"context"
	"io"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// DecodeRows eagerly decodes CSV text into an array of row arrays.
func DecodeRows(ctx context.Context, data runtime.String, opts Options) (runtime.Value, error) {
	iter, err := NewDecodeRowsIterator(data, opts)
	if err != nil {
		return nil, err
	}

	out := runtime.NewArray(10)

	for {
		val, _, err := iter.Next(ctx)
		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, err
		}

		if err := out.Append(ctx, val); err != nil {
			return nil, err
		}
	}

	return out, nil
}
