package lib

import (
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// RegisterLib registers the CSV::DECODE, CSV::DECODE_ROWS,
// CSV::DECODE_STREAM, CSV::DECODE_ROWS_STREAM, and CSV::ENCODE functions in
// the provided namespace.
func RegisterLib(ns runtime.Namespace) error {
	return sdk.RegisterFunctions(
		ns,
		sdk.Func("DECODE", Decode),
		sdk.Func("DECODE_ROWS", DecodeRows),
		sdk.Func("DECODE_STREAM", DecodeStream),
		sdk.Func("DECODE_ROWS_STREAM", DecodeRowsStream),
		sdk.Func("ENCODE", Encode),
	)
}
