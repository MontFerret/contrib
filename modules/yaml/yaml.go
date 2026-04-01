package yaml

import (
	"github.com/MontFerret/contrib/modules/yaml/lib"
	"github.com/MontFerret/ferret/v2"
)

type module struct {
}

// New returns the YAML module, which registers the YAML namespace functions on
// a Ferret host during bootstrap.
func New() ferret.Module {
	return &module{}
}

func (m *module) Name() string {
	return "yaml"
}

func (m *module) Register(bootstrap ferret.Bootstrap) error {
	lib.RegisterLib(bootstrap.Host().Library().Namespace("YAML"))

	return nil
}
