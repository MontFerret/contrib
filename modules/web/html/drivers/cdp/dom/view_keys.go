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
		if r == '-' || r == '_' {
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
