package types

import (
	"errors"
	"fmt"
)

// ErrHeaderColumnConflict reports that both a header row and explicit columns
// were requested at the same time.
var ErrHeaderColumnConflict = errors.New("csv: cannot specify both header: true and columns option")

// ResolveHeaders returns the headers to use for object decoding, whether the
// first row was consumed as a header row, and any validation error.
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
	baseCounts := make(map[string]int, len(headers))
	emitted := make(map[string]struct{}, len(headers))

	for i, h := range headers {
		if h == "" {
			if opts.Strict {
				return nil, newCSVErrorf(1, "empty header name at column %d", i+1)
			}

			h = fmt.Sprintf("col%d", i+1)
		}

		base := h

		if count, exists := baseCounts[base]; exists {
			if opts.Strict {
				return nil, newCSVErrorf(1, "duplicate header name %q", base)
			}

			next := count + 1
			for {
				candidate := fmt.Sprintf("%s_%d", base, next)
				if _, exists := emitted[candidate]; !exists {
					h = candidate
					baseCounts[base] = next
					break
				}

				next++
			}
		} else {
			baseCounts[base] = 1
		}

		emitted[h] = struct{}{}
		result[i] = h
	}

	return result, nil
}
