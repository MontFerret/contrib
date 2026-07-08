package lib

import (
	"github.com/MontFerret/contrib/modules/document/pdf/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// RegisterLib registers the DOCUMENT::PDF namespace functions in the provided
// namespace.
func RegisterLib(ns runtime.Namespace, options ...core.OpenOptions) {
	openOptions := core.DefaultOpenOptions()
	if len(options) > 0 {
		openOptions = options[0]
	}

	ns.Function().A1().
		Add("OPEN", openWithOptions(openOptions))
}
