package types

import (
	"strconv"
	"strings"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func ConvertValue(raw string, opts Options) runtime.Value {
	if opts.Trim {
		raw = strings.TrimSpace(raw)
	}

	if len(opts.NullValues) > 0 {
		for _, nv := range opts.NullValues {
			if raw == nv {
				return runtime.None
			}
		}
	}

	if !opts.InferTypes {
		return runtime.NewString(raw)
	}

	// Try integer
	if i, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return runtime.NewInt(int(i))
	}

	// Try float
	if f, err := strconv.ParseFloat(raw, 64); err == nil {
		return runtime.NewFloat(f)
	}

	// Try boolean
	lower := strings.ToLower(raw)
	if lower == "true" {
		return runtime.True
	}

	if lower == "false" {
		return runtime.False
	}

	return runtime.NewString(raw)
}
