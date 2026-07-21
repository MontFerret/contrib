package lib

import (
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// RegisterLib registers the YAML namespace functions in the provided
// namespace.
func RegisterLib(ns runtime.Namespace) error {
	return sdk.RegisterFunctions(
		ns,
		sdk.Func("DECODE", Decode),
		sdk.Func("DECODE_ALL", DecodeAll),
		sdk.Func("ENCODE", Encode),
	)
}
