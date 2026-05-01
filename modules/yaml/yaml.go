package yaml

import (
	"github.com/MontFerret/contrib/modules/yaml/lib"
	"github.com/MontFerret/ferret/v2/pkg/module"
)

type mod struct {
}

// New returns the YAML module, which registers the YAML namespace functions on
// a Ferret host during bootstrap.
func New() module.Module {
	return &mod{}
}

func (m *mod) Name() string {
	return "yaml"
}

func (m *mod) Register(bootstrap module.Bootstrap) error {
	lib.RegisterLib(bootstrap.Host().Library().Namespace("YAML"))

	return nil
}
