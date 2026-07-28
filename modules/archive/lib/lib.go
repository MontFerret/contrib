package lib

import (
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// RegisterLib registers ARCHIVE namespace functions.
func RegisterLib(ns runtime.Namespace) error {
	return sdk.RegisterFunctions(
		ns,
		sdk.Func("LIST", List),
		sdk.Func("READ", Read),
		sdk.Func("EXTRACT", Extract),
	)
}
