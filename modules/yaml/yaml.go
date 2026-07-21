package yaml

import (
	"github.com/MontFerret/contrib/modules/yaml/lib"
	"github.com/MontFerret/ferret/v2/pkg/module"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// New returns the YAML module, which registers the YAML namespace functions on
// a Ferret host during bootstrap.
func New() module.Module {
	return sdk.NewModule("yaml", func(bootstrap module.Bootstrap) error {
		return lib.RegisterLib(bootstrap.Host().Library().Namespace("YAML"))
	})
}
