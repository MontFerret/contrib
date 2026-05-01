package article

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/article/core"
	"github.com/MontFerret/contrib/modules/web/article/lib"
	"github.com/MontFerret/ferret/v2/pkg/module"
)

type mod struct{}

// New returns the WEB::ARTICLE module, which registers the WEB::ARTICLE
// namespace functions on a Ferret host during bootstrap.
func New() module.Module {
	return &mod{}
}

func (m *mod) Name() string {
	return "web/article"
}

func (m *mod) Register(bootstrap module.Bootstrap) error {
	lib.RegisterLib(bootstrap.Host().Library().Namespace("WEB").Namespace("ARTICLE"))
	bootstrap.Hooks().Session().BeforeRun(func(ctx context.Context) (context.Context, error) {
		return core.WithExtractor(ctx, core.NewExtractor()), nil
	})

	return nil
}
