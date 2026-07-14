package lib

import (
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// RegisterLib registers the DOCUMENT::XLSX namespace functions in the provided
// namespace.
func RegisterLib(ns runtime.Namespace) error {
	return sdk.RegisterFunctions(
		ns,
		sdk.Func("CREATE", Create),
		sdk.Func("OPEN", Open),
		sdk.Func("SHEETS", Sheets),
		sdk.Func("SAVE", Save),
		sdk.Func("CLOSE", Close),
		sdk.Func("SHEET", Sheet),
		sdk.Func("ADD_SHEET", AddSheet),
		sdk.Func("DELETE_SHEET", DeleteSheet),
		sdk.Func("GET", Get),
		sdk.Func("RANGE", Range),
		sdk.Func("APPEND", Append),
		sdk.Func("SAVE_AS", SaveAs),
		sdk.Func("SET", Set),
		sdk.Func("WRITE_RANGE", WriteRange),
	)
}
