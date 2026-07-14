package lib

import (
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// RegisterLib registers the WEB::SITEMAP namespace functions.
func RegisterLib(ns runtime.Namespace) error {
	return sdk.RegisterFunctions(
		ns,
		sdk.Func("FETCH", Fetch),
		sdk.Func("URLS", URLs),
		sdk.Func("STREAM", Stream),
	)
}
