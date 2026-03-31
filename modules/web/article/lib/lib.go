package lib

import "github.com/MontFerret/ferret/v2/pkg/runtime"

// RegisterLib registers the WEB::ARTICLE namespace functions.
func RegisterLib(ns runtime.Namespace) {
	ns.Function().Var().
		Add("EXTRACT", Extract).
		Add("TEXT", Text).
		Add("MARKDOWN", Markdown)
}
