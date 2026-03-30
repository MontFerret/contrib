package lib

import "github.com/MontFerret/ferret/v2/pkg/runtime"

// RegisterLib registers the WEB::ROBOTS namespace functions.
func RegisterLib(ns runtime.Namespace) {
	ns.Function().Var().
		Add("PARSE", Parse).
		Add("ALLOWS", Allows).
		Add("MATCH", Match).
		Add("SITEMAPS", Sitemaps)
}
