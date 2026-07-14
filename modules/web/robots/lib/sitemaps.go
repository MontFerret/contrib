package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/robots/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// Sitemaps returns the declared sitemap URLs from a parsed robots object.
func Sitemaps(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 1); err != nil {
		return nil, err
	}

	doc, err := sdk.DecodeArg[core.Document](ctx, args, 0)
	if err != nil {
		return nil, err
	}

	return sdk.Encode(ctx, core.SitemapValues(doc))
}
