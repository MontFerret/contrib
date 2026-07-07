package lib

import "github.com/MontFerret/ferret/v2/pkg/runtime"

// RegisterLib registers the DOCUMENT::XLSX namespace functions in the provided
// namespace.
func RegisterLib(ns runtime.Namespace) {
	ns.Function().A0().
		Add("CREATE", Create)

	ns.Function().A1().
		Add("OPEN", Open).
		Add("SHEETS", Sheets).
		Add("SAVE", Save).
		Add("CLOSE", Close)

	ns.Function().A2().
		Add("SHEET", Sheet).
		Add("ADD_SHEET", AddSheet).
		Add("DELETE_SHEET", DeleteSheet).
		Add("GET", Get).
		Add("RANGE", Range).
		Add("APPEND", Append).
		Add("SAVE_AS", SaveAs)

	ns.Function().A3().
		Add("SET", Set).
		Add("WRITE_RANGE", WriteRange)
}
