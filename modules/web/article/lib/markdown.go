package lib

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Markdown returns only the cleaned article Markdown.
func Markdown(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	article, err := extractArticle(ctx, args...)
	if err != nil {
		return nil, err
	}

	if article.Markdown == nil {
		return runtime.None, nil
	}

	return runtime.NewString(*article.Markdown), nil
}
