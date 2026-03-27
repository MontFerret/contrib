package types

import (
	"context"
	"io"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func DecodeRows(ctx context.Context, data runtime.String, opts Options) (runtime.Value, error) {
	iter := NewDecodeRowsIterator(data, opts)
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
