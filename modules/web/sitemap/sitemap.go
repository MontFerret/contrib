package sitemap

import (
	"github.com/MontFerret/contrib/modules/web/sitemap/lib"
	"github.com/MontFerret/ferret/v2/pkg/module"
)

type mod struct {
}

// New returns the WEB::SITEMAP module, which registers the WEB::SITEMAP
// namespace functions on a Ferret host during bootstrap.
func New() module.Module {
	return &mod{}
}

func (m *mod) Name() string {
	return "web/sitemap"
}

func (m *mod) Register(bootstrap module.Bootstrap) error {
	lib.RegisterLib(bootstrap.Host().Library().Namespace("WEB").Namespace("SITEMAP"))

	return nil
}
