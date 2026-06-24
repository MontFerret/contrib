package lib

import "github.com/MontFerret/ferret/v2/pkg/runtime"

// RegisterLib registers the REST namespace functions in the provided namespace.
func RegisterLib(ns runtime.Namespace) {
	ns.Function().A1().
		Add("CLIENT", Client)
}
