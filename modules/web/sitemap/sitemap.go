package sitemap

import (
	"github.com/MontFerret/contrib/modules/web/sitemap/lib"
	"github.com/MontFerret/ferret/v2"
)

type module struct {
}

// New returns the WEB module, which registers WEB namespace helpers on a
// Ferret host during bootstrap.
func New() (ferret.Module, error) {
	return &module{}, nil
}

func (m *module) Name() string {
	return "web/sitemap"
}

func (m *module) Register(bootstrap ferret.Bootstrap) error {
	lib.RegisterLib(bootstrap.Host().Library().Namespace("WEB").Namespace("SITEMAP"))

	return nil
}
