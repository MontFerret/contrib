package postgres

import (
	"github.com/MontFerret/contrib/modules/db/postgres/lib"
	"github.com/MontFerret/ferret/v2/pkg/module"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// New returns the DB::POSTGRES module, which registers lifecycle helpers for
// Postgres database handles on a Ferret host during bootstrap.
func New() module.Module {
	return sdk.NewModule("db/postgres", func(bootstrap module.Bootstrap) error {
		return lib.RegisterLib(bootstrap.Host().Library().Namespace("DB").Namespace("POSTGRES"))
	})
}
