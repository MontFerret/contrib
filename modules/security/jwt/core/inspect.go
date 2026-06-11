package core

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Inspect parses a compact JWT without verifying its signature.
func Inspect(ctx context.Context, cfg Config, token runtime.String) (runtime.Value, error) {
	parsed, err := parseCompactToken(token.String(), cfg.maxTokenSize())
	if err != nil {
		return nil, err
	}

	return buildInspectResult(parsed)
}
