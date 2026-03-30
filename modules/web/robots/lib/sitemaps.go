package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/robots/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Sitemaps returns the declared sitemap URLs from a parsed robots object.
func Sitemaps(_ context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 1); err != nil {
		return nil, err
	}

	doc, err := decodeDocument(args[0])
	if err != nil {
		return nil, err
	}

	return encodeValue(core.SitemapValues(doc)), nil
}
