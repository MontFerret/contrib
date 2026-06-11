package core

import "github.com/MontFerret/ferret/v2/pkg/runtime"

// ResolveToken normalizes supported JWT token input to a runtime.String.
func ResolveToken(input runtime.Value) (runtime.String, error) {
	switch content := input.(type) {
	case runtime.String:
		return content, nil
	case runtime.Binary:
		return runtime.NewString(content.String()), nil
	default:
		return runtime.EmptyString, runtime.TypeErrorOf(input, runtime.TypeString, runtime.TypeBinary)
	}
}

// ResolveSecret normalizes supported secret input to bytes.
func ResolveSecret(input runtime.Value) ([]byte, error) {
	switch content := input.(type) {
	case runtime.String:
		return []byte(content.String()), nil
	case runtime.Binary:
		return []byte(content.String()), nil
	default:
		return nil, runtime.TypeErrorOf(input, runtime.TypeString, runtime.TypeBinary)
	}
}
