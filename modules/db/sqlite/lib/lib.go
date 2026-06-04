package lib

import "github.com/MontFerret/ferret/v2/pkg/runtime"

// RegisterLib registers the DB::SQLITE namespace functions in the provided
// namespace.
func RegisterLib(ns runtime.Namespace) {
	ns.Function().Var().
		Add("OPEN", Open).
		Add("CLOSE", Close).
		Add("BEGIN", Begin).
		Add("COMMIT", Commit).
		Add("ROLLBACK", Rollback)
}
