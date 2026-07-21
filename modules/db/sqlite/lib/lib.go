package lib

import (
	"github.com/MontFerret/contrib/modules/db/sqlite/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// RegisterLib registers the DB::SQLITE namespace functions in the provided
// namespace.
func RegisterLib(ns runtime.Namespace, policies ...core.OpenPolicy) error {
	policy := core.DefaultOpenPolicy()

	if len(policies) > 0 {
		policy = policies[0]
	}

	return sdk.RegisterFunctions(
		ns,
		sdk.Func("OPEN", openWithPolicy(policy)),
		sdk.Func("CLOSE", Close),
		sdk.Func("BEGIN", Begin),
		sdk.Func("COMMIT", Commit),
		sdk.Func("ROLLBACK", Rollback),
	)
}
