package toml

import (
	"github.com/MontFerret/contrib/modules/toml/lib"
	"github.com/MontFerret/ferret/v2/pkg/module"
)

type mod struct {
}

// New returns the TOML module, which registers the TOML namespace functions on
// a Ferret host during bootstrap.
func New() module.Module {
	return &mod{}
}

func (m *mod) Name() string {
	return "toml"
}

func (m *mod) Register(bootstrap module.Bootstrap) error {
	lib.RegisterLib(bootstrap.Host().Library().Namespace("TOML"))

	return nil
}
