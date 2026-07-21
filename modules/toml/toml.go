package toml

import (
	"github.com/MontFerret/contrib/modules/toml/lib"
	"github.com/MontFerret/ferret/v2/pkg/module"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// New returns the TOML module, which registers the TOML namespace functions on
// a Ferret host during bootstrap.
func New() module.Module {
	return sdk.NewModule("toml", func(bootstrap module.Bootstrap) error {
		return lib.RegisterLib(bootstrap.Host().Library().Namespace("TOML"))
	})
}
