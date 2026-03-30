package core

import (
	"context"
	"errors"
	"time"

	burnttoml "github.com/BurntSushi/toml"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Decode eagerly decodes a single TOML document into a Ferret runtime object.
func Decode(ctx context.Context, data runtime.String, opts DecodeOptions) (runtime.Value, error) {
	if !opts.Strict {
		return nil, newTOMLError(`decode option "strict=false" is not implemented yet`)
	}

	decoded := make(map[string]any)

	if _, err := burnttoml.Decode(data.String(), &decoded); err != nil {
		return nil, wrapTOMLError(err, "invalid TOML document")
	}

	value, err := normalizeValue(ctx, decoded, opts)
	if err != nil {
		return nil, err
	}

	obj, ok := value.(*runtime.Object)
	if !ok {
		return nil, newTOMLError("top-level TOML document must decode to an object")
	}

	return obj, nil
}

func normalizeValue(ctx context.Context, input any, opts DecodeOptions) (runtime.Value, error) {
	switch value := input.(type) {
	case time.Time:
		return decodeTemporalValue(value, opts)
	case map[string]any:
		return normalizeMap(ctx, value, opts)
	case []map[string]any:
		out := runtime.NewArray(len(value))

		for idx, item := range value {
			normalized, err := normalizeMap(ctx, item, opts)
			if err != nil {
				return nil, runtime.Errorf(err, "at index %d", idx)
			}

			if err := out.Append(ctx, normalized); err != nil {
				return nil, err
			}
		}

		return out, nil
	case []any:
		out := runtime.NewArray(len(value))

		for idx, item := range value {
			normalized, err := normalizeValue(ctx, item, opts)
			if err != nil {
				return nil, runtime.Errorf(err, "at index %d", idx)
			}

			if err := out.Append(ctx, normalized); err != nil {
				return nil, err
			}
		}

		return out, nil
	case nil:
		return nil, newTOMLError("TOML null values are not supported")
	default:
		return normalizeScalarValue(input)
	}
}

func normalizeScalarValue(input any) (runtime.Value, error) {
	value, err := runtime.ValueOf(input)
	if err != nil {
		if errors.Is(err, runtime.ErrRange) {
			return nil, newTOMLErrorf("invalid TOML integer %d exceeds Ferret int range", input)
		}

		return nil, wrapTOMLError(err, "unsupported TOML scalar value")
	}

	switch value.(type) {
	case runtime.String, runtime.Boolean, runtime.Int, runtime.Float:
		return value, nil
	default:
		return nil, newTOMLErrorf("unsupported TOML value type %T", input)
	}
}

func normalizeMap(ctx context.Context, input map[string]any, opts DecodeOptions) (*runtime.Object, error) {
	out := runtime.NewObjectOf(len(input))

	for key, rawValue := range input {
		value, err := normalizeValue(ctx, rawValue, opts)
		if err != nil {
			return nil, runtime.Errorf(err, "at key %q", key)
		}

		if err := out.Set(ctx, runtime.NewString(key), value); err != nil {
			return nil, err
		}
	}

	return out, nil
}
