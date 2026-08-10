package lib

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Extract returns normalized article content and metadata for the provided HTML.
//
// @param input {String|HTMLPage|HTMLDocument|HTMLElement} HTML content or node.
// @return {Object} Normalized article content and metadata.
func Extract(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	article, err := extractArticle(ctx, args...)
	if err != nil {
		return nil, err
	}

	return article.ToValue(), nil
}
