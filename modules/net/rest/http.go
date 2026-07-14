package rest

import (
	"github.com/MontFerret/contrib/modules/net/rest/lib"
	"github.com/MontFerret/ferret/v2/pkg/module"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// New returns the NET::REST module, which registers REST API client helpers on
// a Ferret host during bootstrap.
func New() module.Module {
	return sdk.NewModule("net/rest", func(bootstrap module.Bootstrap) error {
		return lib.RegisterLib(bootstrap.Host().Library().Namespace("NET").Namespace("REST"))
	})
}
