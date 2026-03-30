package lib

import "github.com/MontFerret/ferret/v2/pkg/runtime"

// RegisterLib registers the canonical TOML namespace functions and uppercase
// compatibility aliases in the provided library.
func RegisterLib(lib runtime.Library) {
	lib.Namespace("TOML").Function().Var().
		Add("DECODE", Decode).
		Add("ENCODE", Encode)
}
