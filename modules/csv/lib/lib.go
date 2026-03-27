package lib

import "github.com/MontFerret/ferret/v2/pkg/runtime"

func RegisterLib(ns runtime.Namespace) {
	ns.Function().Var().
		Add("DECODE", Decode).
		Add("DECODE_ROWS", DecodeRows).
		Add("DECODE_STREAM", DecodeStream).
		Add("DECODE_ROWS_STREAM", DecodeRowsStream).
		Add("ENCODE", Encode)
}
