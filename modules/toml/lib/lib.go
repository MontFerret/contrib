package lib

import "github.com/MontFerret/ferret/v2/pkg/runtime"

// RegisterLib registers the canonical TOML namespace functions in the
// provided library.
func RegisterLib(ns runtime.Namespace) {
	ns.Function().Var().
		Add("DECODE", Decode).
		Add("ENCODE", Encode)
}
