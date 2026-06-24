package content

import "github.com/MontFerret/ferret/v2/pkg/runtime"

// StringOrBinary normalizes String and Binary values to a Ferret string.
func StringOrBinary(input runtime.Value) (runtime.String, error) {
	switch value := input.(type) {
	case runtime.String:
		return value, nil
	case runtime.Binary:
		return runtime.NewString(value.String()), nil
	default:
		return runtime.EmptyString, runtime.TypeErrorOf(input, runtime.TypeString, runtime.TypeBinary)
	}
}

// BytesFromStringOrBinary normalizes String and Binary values to bytes.
func BytesFromStringOrBinary(input runtime.Value) ([]byte, error) {
	switch value := input.(type) {
	case runtime.String:
		return []byte(value.String()), nil
	case runtime.Binary:
		return []byte(value), nil
	default:
		return nil, runtime.TypeErrorOf(input, runtime.TypeString, runtime.TypeBinary)
	}
}
