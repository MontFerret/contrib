package rest

import (
	"github.com/MontFerret/contrib/modules/net/rest/lib"
	"github.com/MontFerret/ferret/v2/pkg/module"
)

type mod struct{}

// New returns the NET::REST module, which registers REST API client helpers on
// a Ferret host during bootstrap.
func New() module.Module {
	return &mod{}
}

func (m *mod) Name() string {
	return "net/rest"
}

func (m *mod) Register(bootstrap module.Bootstrap) error {
	lib.RegisterLib(bootstrap.Host().Library().Namespace("NET").Namespace("REST"))

	return nil
}
