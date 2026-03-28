package lib

import "github.com/MontFerret/ferret/v2/pkg/runtime"

// RegisterLib registers the CSV::DECODE, CSV::DECODE_ROWS,
// CSV::DECODE_STREAM, CSV::DECODE_ROWS_STREAM, and CSV::ENCODE functions in
// the provided namespace.
func RegisterLib(ns runtime.Namespace) {
	ns.Function().Var().
		Add("DECODE", Decode).
		Add("DECODE_ROWS", DecodeRows).
		Add("DECODE_STREAM", DecodeStream).
		Add("DECODE_ROWS_STREAM", DecodeRowsStream).
		Add("ENCODE", Encode)
}
