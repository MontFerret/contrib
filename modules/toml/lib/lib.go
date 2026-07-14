package lib

import (
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// RegisterLib registers the canonical TOML namespace functions in the
// provided namespace.
func RegisterLib(ns runtime.Namespace) error {
	return sdk.RegisterFunctions(ns, sdk.Func("DECODE", Decode), sdk.Func("ENCODE", Encode))
}
