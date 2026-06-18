package core

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func runtimeToNative(ctx context.Context, value runtime.Value) (any, error) {
	if runtime.TypeNone.Is(value) {
		return nil, nil
	}

	switch typed := value.(type) {
	case runtime.String:
		return typed.String(), nil
	case runtime.Int:
		return int64(typed), nil
	case runtime.Float:
		return float64(typed), nil
	case runtime.Boolean:
		return bool(typed), nil
	case runtime.Binary:
		return []byte(typed), nil
	case runtime.DateTime:
		return typed.Time.Format(time.RFC3339Nano), nil
	case runtime.Map:
		return mapToNative(ctx, typed)
	case runtime.List:
		return listToNative(ctx, typed)
	default:
		return nil, fmt.Errorf("unsupported runtime value type %T", value)
	}
}

func mapToNative(ctx context.Context, input runtime.Map) (map[string]any, error) {
	out := make(map[string]any)

	err := input.ForEach(ctx, func(ctx context.Context, value, key runtime.Value) (runtime.Boolean, error) {
		converted, err := runtimeToNative(ctx, value)
		if err != nil {
			return runtime.False, fmt.Errorf("at key %q: %w", key.String(), err)
		}

		out[key.String()] = converted

		return runtime.True, nil
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}

func listToNative(ctx context.Context, input runtime.List) ([]any, error) {
	out := make([]any, 0)
	iter, err := input.Iterate(ctx)
	if err != nil {
		return nil, err
	}

	for {
		value, key, err := iter.Next(ctx)
		if errors.Is(err, io.EOF) {
			return out, nil
		}
		if err != nil {
			return nil, err
		}

		converted, err := runtimeToNative(ctx, value)
		if err != nil {
			return nil, fmt.Errorf("at index %s: %w", key.String(), err)
		}

		out = append(out, converted)
	}
}
