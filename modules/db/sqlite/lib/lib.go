package lib

import (
	"github.com/MontFerret/contrib/modules/db/sqlite/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// RegisterLib registers the DB::SQLITE namespace functions in the provided
// namespace.
func RegisterLib(ns runtime.Namespace, policies ...core.OpenPolicy) {
	policy := core.DefaultOpenPolicy()

	if len(policies) > 0 {
		policy = policies[0]
	}

	ns.Function().Var().
		Add("OPEN", openWithPolicy(policy)).
		Add("CLOSE", Close).
		Add("BEGIN", Begin).
		Add("COMMIT", Commit).
		Add("ROLLBACK", Rollback)
}
