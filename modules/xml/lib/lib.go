package lib

import "github.com/MontFerret/ferret/v2/pkg/runtime"

// RegisterLib registers the XML namespace functions in the provided namespace.
func RegisterLib(ns runtime.Namespace) {
	ns.Function().A1().
		Add("ROOT", Root).
		Add("TEXT", Text).
		Add("CHILDREN", Children)

	ns.Function().A2().
		Add("ATTR", Attr)

	ns.Function().Var().
		Add("DECODE", Decode).
		Add("DECODE_STREAM", DecodeStream).
		Add("ENCODE", Encode)
}
