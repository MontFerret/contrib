package lib

import (
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// RegisterLib registers the XML namespace functions in the provided namespace.
func RegisterLib(ns runtime.Namespace) error {
	return sdk.RegisterFunctions(
		ns,
		sdk.Func("ROOT", Root),
		sdk.Func("TEXT", Text),
		sdk.Func("CHILDREN", Children),
		sdk.Func("ATTR", Attr),
		sdk.Func("DECODE", Decode),
		sdk.Func("DECODE_STREAM", DecodeStream),
		sdk.Func("ENCODE", Encode),
	)
}
