package article

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/article/core"
	"github.com/MontFerret/contrib/modules/web/article/lib"
	"github.com/MontFerret/ferret/v2/pkg/module"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// New returns the WEB::ARTICLE module, which registers the WEB::ARTICLE
// namespace functions on a Ferret host during bootstrap.
func New() module.Module {
	return sdk.NewModule("web/article", func(bootstrap module.Bootstrap) error {
		if err := lib.RegisterLib(bootstrap.Host().Library().Namespace("WEB").Namespace("ARTICLE")); err != nil {
			return err
		}

		bootstrap.Hooks().Session().BeforeRun(func(ctx context.Context) (context.Context, error) {
			return core.WithExtractor(ctx, core.NewExtractor()), nil
		})

		return nil
	})
}
