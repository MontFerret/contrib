package types

import (
	"bytes"
	"context"
	"encoding/csv"
	"io"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type DecodeRowsIterator struct {
	reader       *csv.Reader
	opts         Options
	rowNum       runtime.Int
	expectedCols int
}

func NewDecodeRowsIterator(data runtime.String, opts Options) *DecodeRowsIterator {
	reader := csv.NewReader(bytes.NewBufferString(data.String()))
	opts.ApplyToReader(reader)

	return &DecodeRowsIterator{
		reader:       reader,
		opts:         opts,
		rowNum:       0,
		expectedCols: -1,
	}
}

func (d *DecodeRowsIterator) Iterate(_ context.Context) (runtime.Iterator, error) {
	return d, nil
}

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
				return runtime.None, runtime.None, newCSVErrorf(int(d.rowNum), "expected %d fields but got %d", d.expectedCols, len(record))
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
