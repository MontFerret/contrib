package lib

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Extract returns the normalized article object for the provided HTML.
func Extract(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	article, err := extractArticle(ctx, args...)
	if err != nil {
		return nil, err
	}

	return article.ToValue(), nil
}
