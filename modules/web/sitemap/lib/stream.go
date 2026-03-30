package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/sitemap/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// Stream fetches a sitemap tree and returns a lazy URL iterator.
func Stream(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
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

	return sdk.NewProxy(core.NewURLIterator(target.String(), opts)), nil
}
