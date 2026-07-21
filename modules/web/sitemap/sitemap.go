package sitemap

import (
	"github.com/MontFerret/contrib/modules/web/sitemap/lib"
	"github.com/MontFerret/ferret/v2/pkg/module"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// New returns the WEB::SITEMAP module, which registers the WEB::SITEMAP
// namespace functions on a Ferret host during bootstrap.
func New() module.Module {
	return sdk.NewModule("web/sitemap", func(bootstrap module.Bootstrap) error {
		return lib.RegisterLib(bootstrap.Host().Library().Namespace("WEB").Namespace("SITEMAP"))
	})
}
