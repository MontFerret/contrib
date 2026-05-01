package robots

import (
	"github.com/MontFerret/contrib/modules/web/robots/lib"
	"github.com/MontFerret/ferret/v2/pkg/module"
)

type mod struct{}

// New returns the WEB::ROBOTS module, which registers the WEB::ROBOTS
// namespace functions on a Ferret host during bootstrap.
func New() module.Module {
	return &mod{}
}

func (m *mod) Name() string {
	return "web/robots"
}

func (m *mod) Register(bootstrap module.Bootstrap) error {
	lib.RegisterLib(bootstrap.Host().Library().Namespace("WEB").Namespace("ROBOTS"))

	return nil
}
