package core

import (
	"context"
	"fmt"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

var paramsKey = runtime.NewString("params")

func parseParams(ctx context.Context, input runtime.Value) ([]any, error) {
	if input == nil || input == runtime.None {
		return nil, nil
	}

	paramsMap, ok := input.(runtime.Map)
	if !ok {
		return nil, fmt.Errorf("query WITH must be an object")
	}

	if err := paramsMap.ForEach(ctx, func(_ context.Context, _ runtime.Value, key runtime.Value) (runtime.Boolean, error) {
		name, ok := key.(runtime.String)
		if !ok {
			return runtime.False, fmt.Errorf("query WITH keys must be strings")
		}

		if name != paramsKey {
			return runtime.False, fmt.Errorf("query WITH contains unsupported field %q", name.String())
		}

		return runtime.True, nil
	}); err != nil {
		return nil, err
	}

	paramsValue, found, err := paramsMap.Lookup(ctx, paramsKey)
	if err != nil {
		return nil, err
	}

	if !found {
		return nil, nil
	}

	paramsList, ok := paramsValue.(runtime.List)
	if !ok {
		return nil, fmt.Errorf("query WITH.params must be an array")
	}

	params := make([]any, 0)
	index := 0

	if err := runtime.ForEach(ctx, paramsList, func(_ context.Context, value, _ runtime.Value) (runtime.Boolean, error) {
		param, err := runtimeValueToRedisArg(value)
		if err != nil {
			return runtime.False, fmt.Errorf("query WITH.params[%d]: %w", index, err)
		}

		params = append(params, param)
		index++

		return runtime.True, nil
	}); err != nil {
		return nil, err
	}

	return params, nil
}

func runtimeValueToRedisArg(value runtime.Value) (any, error) {
	if value == nil || value == runtime.None {
		return nil, fmt.Errorf("NONE is not a supported Redis command argument")
	}

	switch val := value.(type) {
	case runtime.Int:
		return int64(val), nil
	case runtime.Float:
		return float64(val), nil
	case runtime.String:
		return val.String(), nil
	case runtime.Boolean:
		return bool(val), nil
	case runtime.Binary:
		out := make([]byte, len(val))
		copy(out, val)

		return out, nil
	default:
		return nil, fmt.Errorf("unsupported Redis argument type %s", runtime.TypeName(runtime.TypeOf(value)))
	}
}
