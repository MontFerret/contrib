package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/sitemap/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Fetch fetches and parses a single sitemap document.
func Fetch(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
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

	document, err := core.Fetch(ctx, target.String(), opts)
	if err != nil {
		return nil, err
	}

	return document.ToValue(), nil
}
