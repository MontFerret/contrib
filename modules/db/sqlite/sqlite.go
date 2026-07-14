package sqlite

import (
	"github.com/MontFerret/contrib/modules/db/sqlite/lib"
	"github.com/MontFerret/ferret/v2/pkg/module"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// New returns the DB::SQLITE module, which registers lifecycle helpers for
// SQLite database handles on a Ferret host during bootstrap.
func New(opts ...Option) module.Module {
	o := newOptions(opts)

	return sdk.NewModule("db/sqlite", func(bootstrap module.Bootstrap) error {
		return lib.RegisterLib(
			bootstrap.Host().Library().Namespace("DB").Namespace("SQLITE"),
			o.openPolicy,
		)
	})
}
