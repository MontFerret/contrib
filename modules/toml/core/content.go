package core

import (
	"github.com/MontFerret/contrib/pkg/common/content"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// ResolveContent normalizes supported TOML content input to a runtime.String.
// It accepts only runtime.String and runtime.Binary values.
func ResolveContent(input runtime.Value) (runtime.String, error) {
	return content.StringOrBinary(input)
}
