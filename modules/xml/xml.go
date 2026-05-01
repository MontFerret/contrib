package xml

import (
	"github.com/MontFerret/contrib/modules/xml/lib"
	"github.com/MontFerret/ferret/v2/pkg/module"
)

type mod struct {
}

// New returns the XML module, which registers the XML namespace functions on a
// Ferret host during bootstrap.
func New() module.Module {
	return &mod{}
}

func (m *mod) Name() string {
	return "xml"
}

func (m *mod) Register(bootstrap module.Bootstrap) error {
	lib.RegisterLib(bootstrap.Host().Library().Namespace("XML"))

	return nil
}
