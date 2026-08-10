package core

import (
	"context"
	"fmt"
	"math"
	"math/big"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func redisValueToRuntime(ctx context.Context, input any) (runtime.Value, error) {
	switch value := input.(type) {
	case nil:
		return runtime.None, nil
	case string:
		return runtime.NewString(value), nil
	case []byte:
		return runtime.NewString(string(value)), nil
	case bool:
		return runtime.NewBoolean(value), nil
	case int:
		return runtime.NewInt(value), nil
	case int64:
		return runtime.NewInt64(value), nil
	case int32:
		return runtime.NewInt64(int64(value)), nil
	case int16:
		return runtime.NewInt64(int64(value)), nil
	case int8:
		return runtime.NewInt64(int64(value)), nil
	case uint:
		return redisUintToRuntime(uint64(value))
	case uint64:
		return redisUintToRuntime(value)
	case uint32:
		return runtime.NewInt64(int64(value)), nil
	case uint16:
		return runtime.NewInt64(int64(value)), nil
	case uint8:
		return runtime.NewInt64(int64(value)), nil
	case float64:
		return runtime.NewFloat(value), nil
	case float32:
		return runtime.NewFloat(float64(value)), nil
	case *big.Int:
		if value == nil {
			return runtime.None, nil
		}
		if !value.IsInt64() {
			return runtime.None, fmt.Errorf("redis big integer %s exceeds Ferret integer range", value.String())
		}

		return runtime.NewInt64(value.Int64()), nil
	case []any:
		return redisSliceToRuntime(ctx, value)
	case map[any]any:
		return redisMapToRuntime(ctx, value)
	case map[string]any:
		return redisStringMapToRuntime(ctx, value)
	case map[string]string:
		converted := make(map[string]any, len(value))
		for key, item := range value {
			converted[key] = item
		}

		return redisStringMapToRuntime(ctx, converted)
	case error:
		return runtime.None, value
	default:
		return runtime.None, fmt.Errorf("unsupported Redis response type %T", input)
	}
}

func redisUintToRuntime(value uint64) (runtime.Value, error) {
	if value > math.MaxInt64 {
		return runtime.None, fmt.Errorf("redis integer %d exceeds Ferret integer range", value)
	}

	return runtime.NewInt64(int64(value)), nil
}

func redisSliceToRuntime(ctx context.Context, input []any) (runtime.Value, error) {
	out := runtime.NewArray(len(input))

	for _, item := range input {
		value, err := redisValueToRuntime(ctx, item)
		if err != nil {
			return runtime.None, err
		}

		if err := out.Append(ctx, value); err != nil {
			return runtime.None, err
		}
	}

	return out, nil
}

func redisMapToRuntime(ctx context.Context, input map[any]any) (runtime.Value, error) {
	out := runtime.NewObjectOf(len(input))

	for key, item := range input {
		name, ok := key.(string)
		if !ok {
			return runtime.None, fmt.Errorf("unsupported Redis map key type %T; expected string", key)
		}

		value, err := redisValueToRuntime(ctx, item)
		if err != nil {
			return runtime.None, err
		}

		if err := out.Set(ctx, runtime.NewString(name), value); err != nil {
			return runtime.None, err
		}
	}

	return out, nil
}

func redisStringMapToRuntime(ctx context.Context, input map[string]any) (runtime.Value, error) {
	out := runtime.NewObjectOf(len(input))

	for name, item := range input {
		value, err := redisValueToRuntime(ctx, item)
		if err != nil {
			return runtime.None, err
		}

		if err := out.Set(ctx, runtime.NewString(name), value); err != nil {
			return runtime.None, err
		}
	}

	return out, nil
}
