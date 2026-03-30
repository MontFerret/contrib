package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/article/core"
	"github.com/MontFerret/contrib/modules/web/article/types"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func extractArticle(_ context.Context, args ...runtime.Value) (types.Article, error) {
	if err := runtime.ValidateArgs(args, 1, 1); err != nil {
		return types.Article{}, err
	}

	html, err := runtime.CastArgAt[runtime.String](args, 0)
	if err != nil {
		return types.Article{}, err
	}

	return core.Extract(html.String()), nil
}
