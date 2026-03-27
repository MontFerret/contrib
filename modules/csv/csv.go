package csv

import (
	"github.com/MontFerret/contrib/modules/csv/lib"
	"github.com/MontFerret/ferret/v2"
)

type module struct {
}

func New() (ferret.Module, error) {
	return &module{}, nil
}

func (m *module) Name() string {
	return "csv"
}

func (m *module) Register(bootstrap ferret.Bootstrap) error {
	lib.RegisterLib(bootstrap.Host().Library().Namespace("CSV"))

	return nil
}
