package postgres

import (
	"github.com/MontFerret/contrib/modules/db/postgres/lib"
	"github.com/MontFerret/ferret/v2/pkg/module"
)

type mod struct{}

// New returns the DB::POSTGRES module, which registers lifecycle helpers for
// Postgres database handles on a Ferret host during bootstrap.
func New() module.Module {
	return &mod{}
}

func (m *mod) Name() string {
	return "db/postgres"
}

func (m *mod) Register(bootstrap module.Bootstrap) error {
	lib.RegisterLib(bootstrap.Host().Library().Namespace("DB").Namespace("POSTGRES"))

	return nil
}
