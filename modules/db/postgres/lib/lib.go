package lib

import "github.com/MontFerret/ferret/v2/pkg/runtime"

// RegisterLib registers the DB::POSTGRES namespace functions in the provided
// namespace.
func RegisterLib(ns runtime.Namespace) {
	ns.Function().A1().
		Add("OPEN", Open).
		Add("CLOSE", Close).
		Add("BEGIN", Begin).
		Add("COMMIT", Commit).
		Add("ROLLBACK", Rollback)
}
