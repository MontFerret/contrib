package csv

import (
	"github.com/MontFerret/contrib/modules/csv/lib"
	"github.com/MontFerret/ferret/v2"
)

type module struct {
}

// New returns the CSV module, which registers the CSV namespace functions on a
// Ferret host during bootstrap.
func New() ferret.Module {
	return &module{}
}

func (m *module) Name() string {
	return "csv"
}

func (m *module) Register(bootstrap ferret.Bootstrap) error {
	lib.RegisterLib(bootstrap.Host().Library().Namespace("CSV"))

	return nil
}
