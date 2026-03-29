package lib

import "github.com/MontFerret/ferret/v2/pkg/runtime"

// RegisterLib registers the XML::DECODE, XML::DECODE_STREAM, and XML::ENCODE
// functions in the provided namespace.
func RegisterLib(ns runtime.Namespace) {
	ns.Function().Var().
		Add("DECODE", Decode).
		Add("DECODE_STREAM", DecodeStream).
		Add("ENCODE", Encode)
}
