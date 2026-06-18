package core

import (
	"context"
	"fmt"
	"net/url"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func resolveRequestURL(ctx context.Context, baseURL, expression string, query runtime.Value) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid baseUrl %q: %w", baseURL, err)
	}
	if base.Scheme == "" || base.Host == "" {
		return "", fmt.Errorf("baseUrl must be an absolute URL")
	}

	resource, err := url.Parse(expression)
	if err != nil {
		return "", fmt.Errorf("invalid resource path %q: %w", expression, err)
	}

	resolved := base.ResolveReference(resource)
	values := resolved.Query()
	if err := appendURLValues(ctx, values, "HTTP query WITH.query", query); err != nil {
		return "", err
	}

	resolved.RawQuery = values.Encode()

	return resolved.String(), nil
}
