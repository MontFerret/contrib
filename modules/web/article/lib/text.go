package lib

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Text returns only the cleaned article text.
func Text(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	article, err := extractArticle(ctx, args...)
	if err != nil {
		return nil, err
	}

	if article.Text == nil {
		return runtime.None, nil
	}

	return runtime.NewString(*article.Text), nil
}
