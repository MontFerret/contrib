package lib

import (
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// RegisterLib registers the WEB::ARTICLE namespace functions.
func RegisterLib(ns runtime.Namespace) error {
	return sdk.RegisterFunctions(
		ns,
		sdk.Func("EXTRACT", Extract),
		sdk.Func("TEXT", Text),
		sdk.Func("MARKDOWN", Markdown),
	)
}
