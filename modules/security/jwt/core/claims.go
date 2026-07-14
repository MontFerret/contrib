package core

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

func runtimeMapToClaims(ctx context.Context, value runtime.Map) (map[string]any, error) {
	out := make(map[string]any)

	err := value.ForEach(ctx, func(ctx context.Context, item, key runtime.Value) (runtime.Boolean, error) {
		name, ok := key.(runtime.String)
		if !ok {
			return false, newError(ErrInvalidToken, "claim key must be a string")
		}

		var converted any

		if err := sdk.Decode(ctx, item, &converted); err != nil {
			return false, err
		}

		out[name.String()] = converted

		return true, nil
	})

	if err != nil {
		return nil, err
	}

	return out, nil
}

func claimAsInt64(value any) (int64, bool) {
	switch typed := value.(type) {
	case float64:
		return int64(typed), true
	case int64:
		return typed, true
	case int:
		return int64(typed), true
	default:
		return 0, false
	}
}

func claimAsString(value any) (string, bool) {
	text, ok := value.(string)

	return text, ok && text != ""
}

func audienceMatches(actual any, expected string) bool {
	if expected == "" {
		return true
	}

	switch typed := actual.(type) {
	case string:
		return typed == expected
	case []any:
		for _, item := range typed {
			if text, ok := item.(string); ok && text == expected {
				return true
			}
		}
	}

	return false
}

func hasClaim(claims map[string]any, name string) bool {
	_, ok := claims[name]

	return ok
}
