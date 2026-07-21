package xml

import (
	"github.com/MontFerret/contrib/modules/xml/lib"
	"github.com/MontFerret/ferret/v2/pkg/module"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// New returns the XML module, which registers the XML namespace functions on a
// Ferret host during bootstrap.
func New() module.Module {
	return sdk.NewModule("xml", func(bootstrap module.Bootstrap) error {
		return lib.RegisterLib(bootstrap.Host().Library().Namespace("XML"))
	})
}
