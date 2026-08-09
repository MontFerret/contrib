package core

import (
	"context"
	"fmt"
	"time"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type ExecutionOptions struct {
	Timeout time.Duration
}

func decodeExecutionOptions(ctx context.Context, input runtime.Value) (ExecutionOptions, error) {
	if input == nil || input == runtime.None {
		return ExecutionOptions{}, nil
	}

	optionsMap, ok := input.(runtime.Map)
	if !ok {
		return ExecutionOptions{}, fmt.Errorf("query OPTIONS must be an object")
	}

	var options ExecutionOptions
	if err := optionsMap.ForEach(ctx, func(ctx context.Context, value, key runtime.Value) (runtime.Boolean, error) {
		name, ok := key.(runtime.String)
		if !ok {
			return runtime.False, fmt.Errorf("query OPTIONS keys must be strings")
		}

		if name.String() != "timeout" {
			return runtime.False, fmt.Errorf("query OPTIONS contains unsupported field %q", name.String())
		}

		switch value.(type) {
		case runtime.Duration, runtime.String, runtime.Int, runtime.Float:
		default:
			return runtime.False, fmt.Errorf("query OPTIONS.timeout must be a duration string or number of milliseconds")
		}

		duration, err := runtime.ToDuration(ctx, value)
		if err != nil {
			return runtime.False, fmt.Errorf("invalid query OPTIONS.timeout: %w", err)
		}
		if duration < 0 {
			return runtime.False, fmt.Errorf("query OPTIONS.timeout must be greater than or equal to 0")
		}

		options.Timeout = time.Duration(duration)

		return runtime.True, nil
	}); err != nil {
		return ExecutionOptions{}, err
	}

	return options, nil
}
