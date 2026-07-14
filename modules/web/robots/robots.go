package robots

import (
	"github.com/MontFerret/contrib/modules/web/robots/lib"
	"github.com/MontFerret/ferret/v2/pkg/module"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// New returns the WEB::ROBOTS module, which registers the WEB::ROBOTS
// namespace functions on a Ferret host during bootstrap.
func New() module.Module {
	return sdk.NewModule("web/robots", func(bootstrap module.Bootstrap) error {
		return lib.RegisterLib(bootstrap.Host().Library().Namespace("WEB").Namespace("ROBOTS"))
	})
}
