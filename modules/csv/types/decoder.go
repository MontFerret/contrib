package types

import (
	"context"
	"io"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Decode eagerly decodes CSV text into an array of objects.
func Decode(ctx context.Context, data runtime.String, opts Options) (runtime.Value, error) {
	iter, err := NewDecodeIterator(data, opts)
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

func rowToObject(ctx context.Context, record []string, headers []string, rowNum int, opts Options) (runtime.Value, error) {
	if opts.SkipEmpty && isEmptyRow(record) {
		return nil, nil
	}

	if opts.Strict && len(record) != len(headers) {
		return nil, newCSVErrorf(rowNum, "expected %d fields but got %d", len(headers), len(record))
	}

	obj := runtime.NewObject()

	for i, h := range headers {
		var val runtime.Value

		if i < len(record) {
			val = ConvertValue(record[i], opts)
		} else {
			val = runtime.None
		}

		if err := obj.Set(ctx, runtime.NewString(h), val); err != nil {
			return nil, err
		}
	}

	return obj, nil
}

func isEmptyRow(record []string) bool {
	for _, field := range record {
		if field != "" {
			return false
		}
	}

	return true
}
