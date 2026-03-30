package article

import (
	"github.com/MontFerret/contrib/modules/web/article/lib"
	"github.com/MontFerret/ferret/v2"
)

type module struct{}

// New returns the WEB::ARTICLE module, which registers the WEB::ARTICLE
// namespace functions on a Ferret host during bootstrap.
func New() (ferret.Module, error) {
	return &module{}, nil
}

func (m *module) Name() string {
	return "web/article"
}

func (m *module) Register(bootstrap ferret.Bootstrap) error {
	lib.RegisterLib(bootstrap.Host().Library().Namespace("WEB").Namespace("ARTICLE"))

	return nil
}
