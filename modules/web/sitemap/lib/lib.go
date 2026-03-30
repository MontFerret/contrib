package lib

import "github.com/MontFerret/ferret/v2/pkg/runtime"

// RegisterLib registers the WEB::SITEMAP namespace functions.
func RegisterLib(ns runtime.Namespace) {
	ns.Function().Var().
		Add("FETCH", Fetch).
		Add("URLS", URLs).
		Add("STREAM", Stream)
}
