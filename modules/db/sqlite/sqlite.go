package sqlite

import (
	"github.com/MontFerret/contrib/modules/db/sqlite/core"
	"github.com/MontFerret/contrib/modules/db/sqlite/lib"
	"github.com/MontFerret/ferret/v2/pkg/module"
)

type mod struct {
	openPolicy core.OpenPolicy
}

// New returns the DB::SQLITE module, which registers lifecycle helpers for
// SQLite database handles on a Ferret host during bootstrap.
func New(opts ...Option) module.Module {
	o := newOptions(opts)

	return &mod{
		openPolicy: o.openPolicy,
	}
}

func (m *mod) Name() string {
	return "db/sqlite"
}

func (m *mod) Register(bootstrap module.Bootstrap) error {
	lib.RegisterLib(bootstrap.Host().Library().Namespace("DB").Namespace("SQLITE"), m.openPolicy)

	return nil
}
