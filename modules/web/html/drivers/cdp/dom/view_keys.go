package dom

import (
	"strings"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func datasetPropertyName(key runtime.Value) runtime.String {
	name := strings.TrimPrefix(key.String(), "data-")
	var out strings.Builder
	upperNext := false

	for _, r := range name {
		if r == '-' {
			upperNext = true
			continue
		}

		if upperNext {
			out.WriteString(strings.ToUpper(string(r)))
			upperNext = false
			continue
		}

		out.WriteRune(r)
	}

	return runtime.NewString(out.String())
}

func stylePropertyName(key runtime.Value) runtime.String {
	name := key.String()
	if strings.Contains(name, "-") {
		return runtime.NewString(name)
	}

	var out strings.Builder
	for _, r := range name {
		if r >= 'A' && r <= 'Z' {
			if out.Len() > 0 {
				out.WriteRune('-')
			}

			out.WriteRune(r + ('a' - 'A'))
			continue
		}

		out.WriteRune(r)
	}

	return runtime.NewString(out.String())
}
