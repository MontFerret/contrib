package types

import (
	"bytes"
	"context"
	"encoding/csv"
	"io"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type DecodeIterator struct {
	reader          *csv.Reader
	headers         []string
	headersConsumed bool
	firstRow        []string
	opts            Options
	rowNum          runtime.Int
}

func NewDecodeIterator(data runtime.String, opts Options) (*DecodeIterator, error) {
	reader := csv.NewReader(bytes.NewBufferString(data.String()))
	opts.ApplyToReader(reader)

	// Read first row to determine headers
	firstRow, err := reader.Read()
	if err != nil {
		if err == io.EOF {
			// Empty input — return an exhausted iterator
			return &DecodeIterator{
				reader:          reader,
				headersConsumed: true,
			}, nil
		}

		return nil, runtime.Error(err, "csv: failed to read first row")
	}

	headers, headersConsumed, err := ResolveHeaders(firstRow, opts)
	if err != nil {
		return nil, err
	}

	var rowNum runtime.Int
	if headersConsumed {
		rowNum = 1 // header was row 1, data starts at row 2
	}

	return &DecodeIterator{
		reader:          reader,
		headers:         headers,
		headersConsumed: headersConsumed,
		firstRow:        firstRow,
		opts:            opts,
		rowNum:          rowNum,
	}, nil
}

func (d *DecodeIterator) Iterate(_ context.Context) (runtime.Iterator, error) {
	return d, nil
}

func (d *DecodeIterator) Next(ctx context.Context) (runtime.Value, runtime.Value, error) {
	// If first row was not consumed as header, yield it as data
	if !d.headersConsumed {
		d.headersConsumed = true
		d.rowNum++

		if !(d.opts.SkipEmpty && isEmptyRow(d.firstRow)) {
			obj, err := rowToObject(ctx, d.firstRow, d.headers, int(d.rowNum), d.opts)
			if err != nil {
				return runtime.None, runtime.None, err
			}

			if obj != nil {
				return obj, d.rowNum, nil
			}
		}
	}

	for {
		d.rowNum++
		record, err := d.reader.Read()

		if err != nil {
			if err == io.EOF {
				return runtime.None, runtime.None, io.EOF
			}

			return runtime.None, runtime.None, runtime.Error(err, "csv: failed to decode CSV data")
		}

		obj, err := rowToObject(ctx, record, d.headers, int(d.rowNum), d.opts)
		if err != nil {
			return runtime.None, runtime.None, err
		}

		if obj != nil {
			return obj, d.rowNum, nil
		}
		// obj == nil means empty row skipped, continue loop
	}
}
