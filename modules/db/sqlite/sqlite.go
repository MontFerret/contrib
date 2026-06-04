package sqlite

import (
	"github.com/MontFerret/contrib/modules/db/sqlite/lib"
	"github.com/MontFerret/ferret/v2/pkg/module"
)

type mod struct {
}

// New returns the DB::SQLITE module, which registers lifecycle helpers for
// SQLite database handles on a Ferret host during bootstrap.
func New() module.Module {
	return &mod{}
}

func (m *mod) Name() string {
	return "db/sqlite"
}

func (m *mod) Register(bootstrap module.Bootstrap) error {
	lib.RegisterLib(bootstrap.Host().Library().Namespace("DB").Namespace("SQLITE"))

	return nil
}
