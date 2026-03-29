package lib

import "github.com/MontFerret/ferret/v2/pkg/runtime"

// RegisterLib registers the YAML namespace functions in the provided
// namespace.
func RegisterLib(ns runtime.Namespace) {
	ns.Function().Var().
		Add("DECODE", Decode).
		Add("DECODE_ALL", DecodeAll).
		Add("ENCODE", Encode)
}
