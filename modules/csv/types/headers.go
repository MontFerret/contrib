package types

import (
	"errors"
	"fmt"
)

var ErrHeaderColumnConflict = errors.New("csv: cannot specify both header: true and columns option")

func ResolveHeaders(firstRow []string, opts Options) ([]string, bool, error) {
	if opts.Header && len(opts.Columns) > 0 {
		return nil, false, ErrHeaderColumnConflict
	}

	if opts.Header {
		headers, err := validateHeaders(firstRow, opts)
		return headers, true, err
	}

	if len(opts.Columns) > 0 {
		return opts.Columns, false, nil
	}

	// Auto-generate col1, col2, ...
	headers := make([]string, len(firstRow))
	for i := range firstRow {
		headers[i] = fmt.Sprintf("col%d", i+1)
	}

	return headers, false, nil
}

func validateHeaders(headers []string, opts Options) ([]string, error) {
	result := make([]string, len(headers))
	seen := make(map[string]int, len(headers))

	for i, h := range headers {
		if h == "" {
			if opts.Strict {
				return nil, newCSVErrorf(1, "empty header name at column %d", i+1)
			}

			h = fmt.Sprintf("col%d", i+1)
		}

		base := h

		if count, exists := seen[base]; exists {
			if opts.Strict {
				return nil, newCSVErrorf(1, "duplicate header name %q", base)
			}

			h = fmt.Sprintf("%s_%d", base, count+1)
		}

		seen[base]++
		result[i] = h
	}

	return result, nil
}
