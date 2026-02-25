package common

import (
	"strings"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func PathToString(path []runtime.Value) string {
	spath := make([]string, 0, len(path))

	for i, s := range path {
		spath[i] = s.String()
	}

	return strings.Join(spath, ".")
}
