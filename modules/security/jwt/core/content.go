package core

import (
	"github.com/MontFerret/contrib/pkg/common/content"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// ResolveToken normalizes supported JWT token input to a runtime.String.
func ResolveToken(input runtime.Value) (runtime.String, error) {
	return content.StringOrBinary(input)
}

// ResolveSecret normalizes supported secret input to bytes.
func ResolveSecret(input runtime.Value) ([]byte, error) {
	return content.BytesFromStringOrBinary(input)
}
