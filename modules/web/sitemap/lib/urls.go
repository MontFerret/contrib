package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/sitemap/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// URLs fetches a sitemap tree and returns flattened URL entries.
func URLs(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 2); err != nil {
		return nil, err
	}

	target, err := runtime.CastArgAt[runtime.String](args, 0)
	if err != nil {
		return nil, err
	}

	opts, err := parseOptions(ctx, args)
	if err != nil {
		return nil, err
	}

	entries, err := core.CollectURLs(ctx, target.String(), opts)
	if err != nil {
		return nil, err
	}

	return core.URLEntriesToValue(entries), nil
}
