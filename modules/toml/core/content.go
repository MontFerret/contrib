package core

import "github.com/MontFerret/ferret/v2/pkg/runtime"

// ResolveContent normalizes supported TOML content input to a runtime.String.
// It accepts only runtime.String and runtime.Binary values.
func ResolveContent(input runtime.Value) (runtime.String, error) {
	switch content := input.(type) {
	case runtime.String:
		return content, nil
	case runtime.Binary:
		return runtime.NewString(content.String()), nil
	default:
		return runtime.EmptyString, runtime.TypeErrorOf(input, runtime.TypeString, runtime.TypeBinary)
	}
}
