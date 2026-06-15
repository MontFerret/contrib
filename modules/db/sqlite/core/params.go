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
		return nil, fmt.Errorf("query params must be an object")
	}

	paramsValue, err := paramsMap.Get(ctx, paramsKey)
	if err != nil {
		return nil, err
	}
	if paramsValue == nil || paramsValue == runtime.None {
		return nil, nil
	}

	paramsList, ok := paramsValue.(runtime.List)
	if !ok {
		return nil, fmt.Errorf("params must be an array")
	}

	params := make([]any, 0)
	if err := runtime.ForEach(ctx, paramsList, func(ctx context.Context, value, _ runtime.Value) (runtime.Boolean, error) {
		param, err := runtimeValueToSQLParam(value)
		if err != nil {
			return runtime.False, err
		}

		params = append(params, param)

		return runtime.True, nil
	}); err != nil {
		return nil, err
	}

	return params, nil
}

func runtimeValueToSQLParam(value runtime.Value) (any, error) {
	if value == nil || value == runtime.None {
		return nil, nil
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
		return nil, fmt.Errorf("unsupported param type %s", runtime.TypeName(runtime.TypeOf(value)))
	}
}
