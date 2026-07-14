package lib

import (
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// RegisterLib registers the WEB::ROBOTS namespace functions.
func RegisterLib(ns runtime.Namespace) error {
	return sdk.RegisterFunctions(
		ns,
		sdk.Func("PARSE", Parse),
		sdk.Func("ALLOWS", Allows),
		sdk.Func("MATCH", Match),
		sdk.Func("SITEMAPS", Sitemaps),
	)
}
