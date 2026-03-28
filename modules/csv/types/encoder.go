package types

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"io"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// EncodeResult contains the encoded CSV text and the number of data rows
// written.
type EncodeResult struct {
	// Text is the encoded CSV output.
	Text string
	// Rows is the number of data rows written, excluding any generated header.
	Rows int
}

// Encode encodes an array of objects or row arrays into CSV text.
func Encode(ctx context.Context, data runtime.Value, opts Options) (*EncodeResult, error) {
	list, ok := data.(runtime.List)
	if !ok {
		return nil, errors.New("csv: encode expects an array")
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	if err := opts.ApplyToWriter(writer); err != nil {
		return nil, err
	}

	// Peek at first element to determine mode
	iter, err := list.Iterate(ctx)
	if err != nil {
		return nil, runtime.Error(err, "csv: failed to iterate data")
	}

	firstVal, _, err := iter.Next(ctx)
	if err != nil {
		if err == io.EOF {
			if err := flushWriter(writer); err != nil {
				return nil, err
			}

			return &EncodeResult{Text: buf.String(), Rows: 0}, nil
		}

		return nil, runtime.Error(err, "csv: failed to read first element")
	}

	_, isMap := firstVal.(runtime.Map)

	if isMap {
		return encodeRecords(ctx, writer, &buf, iter, firstVal, opts)
	}

	return encodeArrays(ctx, writer, &buf, iter, firstVal, opts)
}

func encodeRecords(ctx context.Context, writer *csv.Writer, buf *bytes.Buffer, iter runtime.Iterator, firstVal runtime.Value, opts Options) (*EncodeResult, error) {
	// Determine columns
	columns, err := resolveEncodeColumns(ctx, firstVal.(runtime.Map), opts)
	if err != nil {
		return nil, err
	}

	// Write header if requested
	if opts.Header {
		if err := writer.Write(columns); err != nil {
			return nil, runtime.Error(err, "csv: failed to write header")
		}
	}

	rowCount := 0

	// Write first record
	row, err := mapToRecord(ctx, firstVal.(runtime.Map), columns)
	if err != nil {
		return nil, err
	}

	if err := writer.Write(row); err != nil {
		return nil, runtime.Error(err, "csv: failed to write row")
	}

	rowCount++

	// Write remaining records
	for {
		val, _, err := iter.Next(ctx)
		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, runtime.Error(err, "csv: failed to iterate data")
		}

		m, ok := val.(runtime.Map)
		if !ok {
			return nil, errors.New("csv: encode expected all elements to be objects")
		}

		row, err := mapToRecord(ctx, m, columns)
		if err != nil {
			return nil, err
		}

		if err := writer.Write(row); err != nil {
			return nil, runtime.Error(err, "csv: failed to write row")
		}

		rowCount++
	}

	if err := flushWriter(writer); err != nil {
		return nil, err
	}

	return &EncodeResult{Text: buf.String(), Rows: rowCount}, nil
}

func encodeArrays(ctx context.Context, writer *csv.Writer, buf *bytes.Buffer, iter runtime.Iterator, firstVal runtime.Value, opts Options) (*EncodeResult, error) {
	rowCount := 0

	// Write first row
	row, err := arrayToRecord(ctx, firstVal)
	if err != nil {
		return nil, err
	}

	if err := writer.Write(row); err != nil {
		return nil, runtime.Error(err, "csv: failed to write row")
	}

	rowCount++

	// Write remaining rows
	for {
		val, _, err := iter.Next(ctx)
		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, runtime.Error(err, "csv: failed to iterate data")
		}

		row, err := arrayToRecord(ctx, val)
		if err != nil {
			return nil, err
		}

		if err := writer.Write(row); err != nil {
			return nil, runtime.Error(err, "csv: failed to write row")
		}

		rowCount++
	}

	if err := flushWriter(writer); err != nil {
		return nil, err
	}

	return &EncodeResult{Text: buf.String(), Rows: rowCount}, nil
}

func flushWriter(writer *csv.Writer) error {
	writer.Flush()
	if err := writer.Error(); err != nil {
		return runtime.Error(err, "csv: failed to flush output")
	}

	return nil
}

func resolveEncodeColumns(ctx context.Context, first runtime.Map, opts Options) ([]string, error) {
	if len(opts.Columns) > 0 {
		return opts.Columns, nil
	}

	// Derive from first object's keys
	var columns []string

	err := first.ForEach(ctx, func(_ context.Context, value, key runtime.Value) (runtime.Boolean, error) {
		columns = append(columns, key.String())

		return true, nil
	})

	if err != nil {
		return nil, runtime.Error(err, "csv: failed to read object keys")
	}

	return columns, nil
}

func mapToRecord(ctx context.Context, m runtime.Map, columns []string) ([]string, error) {
	row := make([]string, len(columns))

	for i, col := range columns {
		val, err := m.Get(ctx, runtime.NewString(col))
		if err != nil {
			row[i] = ""

			continue
		}

		if val == runtime.None {
			row[i] = ""
		} else {
			row[i] = val.String()
		}
	}

	return row, nil
}

func arrayToRecord(ctx context.Context, val runtime.Value) ([]string, error) {
	list, ok := val.(runtime.List)
	if !ok {
		return nil, errors.New("csv: encode expected array elements to be arrays")
	}

	iter, err := list.Iterate(ctx)
	if err != nil {
		return nil, runtime.Error(err, "csv: failed to iterate row")
	}

	var row []string

	for {
		v, _, err := iter.Next(ctx)
		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, err
		}

		if v == runtime.None {
			row = append(row, "")
		} else {
			row = append(row, v.String())
		}
	}

	return row, nil
}
