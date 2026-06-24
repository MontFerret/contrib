package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/robots/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// Sitemaps returns the declared sitemap URLs from a parsed robots object.
func Sitemaps(_ context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 1); err != nil {
		return nil, err
	}

	var doc core.Document
	if err := sdk.Decode(args[0], &doc); err != nil {
		return nil, err
	}

	return sdk.Encode(core.SitemapValues(doc)), nil
}
