package core

import (
	"bytes"
	"context"
	"encoding/csv"
	"io"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// DecodeIterator iterates over decoded CSV objects.
// The iterator key is the original 1-based CSV record number after parsing.
type DecodeIterator struct {
	reader          *csv.Reader
	headers         []string
	firstRow        []string
	opts            Options
	rowNum          runtime.Int
	headersConsumed bool
}

// NewDecodeIterator returns an iterator over decoded CSV objects.
func NewDecodeIterator(data runtime.String, opts Options) (*DecodeIterator, error) {
	reader := csv.NewReader(bytes.NewBufferString(data.String()))
	if err := opts.ApplyToReader(reader); err != nil {
		return nil, err
	}

	firstRow, rowNum, err := readFirstDataRow(reader, opts)
	if err != nil {
		return nil, err
	}

	if firstRow == nil {
		return &DecodeIterator{
			reader:          reader,
			headersConsumed: true,
		}, nil
	}

	headers, headersConsumed, err := ResolveHeaders(firstRow, opts)
	if err != nil {
		return nil, err
	}

	if headersConsumed {
		// Header was read from the original CSV record number tracked in rowNum.
		// The next yielded row should continue from that record number.
		return &DecodeIterator{
			reader:          reader,
			headers:         headers,
			headersConsumed: headersConsumed,
			firstRow:        firstRow,
			opts:            opts,
			rowNum:          rowNum,
		}, nil
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

func readFirstDataRow(reader *csv.Reader, opts Options) ([]string, runtime.Int, error) {
	var rowNum runtime.Int

	for {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				return nil, 0, nil
			}

			return nil, 0, runtime.Error(err, "csv: failed to read first row")
		}

		rowNum++

		if opts.SkipEmpty && isEmptyRow(record) {
			continue
		}

		return record, rowNum, nil
	}
}

// Iterate returns the iterator itself.
func (d *DecodeIterator) Iterate(_ context.Context) (runtime.Iterator, error) {
	return d, nil
}

// Next returns the next decoded object and its CSV record number.
func (d *DecodeIterator) Next(ctx context.Context) (runtime.Value, runtime.Value, error) {
	// If first row was not consumed as header, yield it as data
	if !d.headersConsumed {
		d.headersConsumed = true

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
