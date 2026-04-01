package toml

import (
	"github.com/MontFerret/contrib/modules/toml/lib"
	"github.com/MontFerret/ferret/v2"
)

type module struct {
}

// New returns the TOML module, which registers the TOML namespace functions on
// a Ferret host during bootstrap.
func New() ferret.Module {
	return &module{}
}

func (m *module) Name() string {
	return "toml"
}

func (m *module) Register(bootstrap ferret.Bootstrap) error {
	lib.RegisterLib(bootstrap.Host().Library().Namespace("TOML"))

	return nil
}
