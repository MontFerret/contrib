package lib

import (
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// RegisterLib registers the DB::POSTGRES namespace functions in the provided
// namespace.
func RegisterLib(ns runtime.Namespace) error {
	return sdk.RegisterFunctions(
		ns,
		sdk.Func("OPEN", Open),
		sdk.Func("CLOSE", Close),
		sdk.Func("BEGIN", Begin),
		sdk.Func("COMMIT", Commit),
		sdk.Func("ROLLBACK", Rollback),
	)
}
