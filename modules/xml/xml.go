package xml

import (
	"github.com/MontFerret/contrib/modules/xml/lib"
	"github.com/MontFerret/ferret/v2"
)

type module struct {
}

// New returns the XML module, which registers the XML namespace functions on a
// Ferret host during bootstrap.
func New() ferret.Module {
	return &module{}
}

func (m *module) Name() string {
	return "xml"
}

func (m *module) Register(bootstrap ferret.Bootstrap) error {
	lib.RegisterLib(bootstrap.Host().Library().Namespace("XML"))

	return nil
}
