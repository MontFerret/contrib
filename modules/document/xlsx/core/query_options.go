package core

import (
	"context"
	"fmt"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type QueryOptions struct {
	Headers       bool
	TrimEmptyRows bool
}

func decodeQueryOptions(ctx context.Context, value runtime.Value) (QueryOptions, error) {
	var opts QueryOptions

	if value == nil || value == runtime.None {
		return opts, nil
	}

	obj, ok := value.(runtime.Map)
	if !ok {
		return opts, fmt.Errorf("XLSX query WITH must be an object")
	}

	if err := obj.ForEach(ctx, func(_ context.Context, value, key runtime.Value) (runtime.Boolean, error) {
		name, ok := key.(runtime.String)

		if !ok {
			return runtime.False, fmt.Errorf("XLSX query WITH keys must be strings")
		}

		switch name.String() {
		case "headers":
			flag, ok := value.(runtime.Boolean)
			if !ok {
				return runtime.False, fmt.Errorf("XLSX query WITH.%s must be a boolean", name.String())
			}

			opts.Headers = bool(flag)
		case "trimEmptyRows":
			flag, ok := value.(runtime.Boolean)
			if !ok {
				return runtime.False, fmt.Errorf("XLSX query WITH.%s must be a boolean", name.String())
			}

			opts.TrimEmptyRows = bool(flag)
		default:
			return runtime.False, fmt.Errorf("unsupported XLSX query WITH.%s", name.String())
		}

		return runtime.True, nil
	}); err != nil {
		return opts, err
	}

	return opts, nil
}

func applyQueryOptions(ctx context.Context, rows [][]runtime.Value, opts QueryOptions) (runtime.List, error) {
	if opts.Headers {
		return rowsWithHeaders(ctx, rows, opts.TrimEmptyRows)
	}

	if opts.TrimEmptyRows {
		rows = trimEmptyRows(rows)
	}

	return runtimeRowsToArray(ctx, rows)
}

func rowsWithHeaders(ctx context.Context, rows [][]runtime.Value, trim bool) (runtime.List, error) {
	if len(rows) == 0 {
		return runtime.NewArray(0), nil
	}

	headers := normalizeHeaders(rows[0])
	dataRows := rows[1:]

	if trim {
		dataRows = trimEmptyRows(dataRows)
	}

	out := runtime.NewArray(len(dataRows))
	for _, row := range dataRows {
		obj := runtime.NewObjectOf(len(headers))

		for idx, name := range headers {
			if err := obj.Set(ctx, runtime.NewString(name), row[idx]); err != nil {
				return nil, err
			}
		}

		if err := out.Append(ctx, obj); err != nil {
			return nil, err
		}
	}

	return out, nil
}

func normalizeHeaders(row []runtime.Value) []string {
	headers := make([]string, len(row))
	seen := make(map[string]int, len(row))

	for idx, value := range row {
		name := ""

		if value != nil {
			name = value.String()
		}

		if name == "" {
			name = fmt.Sprintf("column_%d", idx+1)
		}

		seen[name]++

		if seen[name] > 1 {
			name = fmt.Sprintf("%s_%d", name, seen[name])
			seen[name]++
		}

		headers[idx] = name
	}

	return headers
}

func trimEmptyRows(rows [][]runtime.Value) [][]runtime.Value {
	end := len(rows)

	for end > 0 && isEmptyRow(rows[end-1]) {
		end--
	}

	return rows[:end]
}

func isEmptyRow(row []runtime.Value) bool {
	for _, value := range row {
		if value != nil && value != runtime.None {
			return false
		}
	}

	return true
}
