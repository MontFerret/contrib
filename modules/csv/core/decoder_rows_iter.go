package core

import (
	"bytes"
	"context"
	"encoding/csv"
	"io"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// DecodeRowsIterator iterates over decoded CSV rows.
// The iterator key is the original 1-based CSV record number after parsing.
type DecodeRowsIterator struct {
	reader       *csv.Reader
	opts         Options
	rowNum       runtime.Int
	expectedCols int
}

// NewDecodeRowsIterator returns an iterator over decoded CSV rows.
func NewDecodeRowsIterator(data runtime.String, opts Options) (*DecodeRowsIterator, error) {
	reader := csv.NewReader(bytes.NewBufferString(data.String()))
	if err := opts.ApplyToReader(reader); err != nil {
		return nil, err
	}

	return &DecodeRowsIterator{
		reader:       reader,
		opts:         opts,
		rowNum:       0,
		expectedCols: -1,
	}, nil
}

// Iterate returns the iterator itself.
func (d *DecodeRowsIterator) Iterate(_ context.Context) (runtime.Iterator, error) {
	return d, nil
}

// Next returns the next decoded row and its CSV record number.
func (d *DecodeRowsIterator) Next(ctx context.Context) (runtime.Value, runtime.Value, error) {
	for {
		d.rowNum++
		record, err := d.reader.Read()

		if err != nil {
			if err == io.EOF {
				return runtime.None, runtime.None, io.EOF
			}

			return runtime.None, runtime.None, runtime.Error(err, "csv: failed to decode CSV data")
		}

		if d.opts.SkipEmpty && isEmptyRow(record) {
			continue
		}

		if d.opts.Strict {
			if d.expectedCols < 0 {
				d.expectedCols = len(record)
			} else if len(record) != d.expectedCols {
				return runtime.None, runtime.None, newErrorf(int(d.rowNum), "expected %d fields but got %d", d.expectedCols, len(record))
			}
		}

		row := runtime.NewArray(len(record))

		for _, field := range record {
			val := ConvertValue(field, d.opts)

			if err := row.Append(ctx, val); err != nil {
				return runtime.None, runtime.None, err
			}
		}

		return row, d.rowNum, nil
	}
}
