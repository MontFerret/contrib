package core

import (
	"context"

	yaml "github.com/goccy/go-yaml"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Encode serializes a Ferret runtime value into YAML text.
func Encode(ctx context.Context, value runtime.Value) (string, error) {
	native, err := toYAMLValue(ctx, value)
	if err != nil {
		return "", err
	}

	data, err := yaml.Marshal(native)
	if err != nil {
		return "", wrapError(err, "failed to encode YAML output")
	}

	return string(data), nil
}

func toYAMLValue(ctx context.Context, value runtime.Value) (any, error) {
	if value == nil || value == runtime.None {
		return nil, nil
	}

	switch current := value.(type) {
	case runtime.Boolean:
		return bool(current), nil
	case runtime.Int:
		return int64(current), nil
	case runtime.Float:
		return float64(current), nil
	case runtime.String:
		return current.String(), nil
	case *runtime.Array:
		return encodeArray(ctx, current)
	case runtime.Map:
		return encodeMap(ctx, current)
	case runtime.Binary, runtime.DateTime, runtime.Iterator, runtime.Observable, runtime.Queryable:
		return nil, newError("unsupported value type for YAML encoding")
	default:
		return nil, newErrorf("unsupported value type for YAML encoding: %T", value)
	}
}

func encodeArray(ctx context.Context, input *runtime.Array) ([]any, error) {
	length, err := input.Length(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]any, 0, int(length))

	for idx := runtime.Int(0); idx < length; idx++ {
		item, err := input.At(ctx, idx)
		if err != nil {
			return nil, err
		}

		value, err := toYAMLValue(ctx, item)
		if err != nil {
			return nil, runtime.Errorf(err, "at index %d", idx)
		}

		out = append(out, value)
	}

	return out, nil
}

func encodeMap(ctx context.Context, input runtime.Map) (map[string]any, error) {
	out := make(map[string]any)

	if err := input.ForEach(ctx, func(ctx context.Context, value, key runtime.Value) (runtime.Boolean, error) {
		encoded, err := toYAMLValue(ctx, value)
		if err != nil {
			return false, runtime.Errorf(err, "at key %q", key.String())
		}

		out[key.String()] = encoded

		return true, nil
	}); err != nil {
		return nil, err
	}

	return out, nil
}
