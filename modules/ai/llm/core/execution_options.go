package core

import (
	"context"
	"math"
	"time"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// DecodeExecutionOptions validates common provider execution policy.
func DecodeExecutionOptions(ctx context.Context, value runtime.Value) (ExecutionOptions, error) {
	const label = "execution options"
	values, err := optionValues(ctx, value, label)
	if err != nil {
		return ExecutionOptions{}, err
	}

	allowed := map[string]struct{}{"temperature": {}, "maxOutputTokens": {}, "timeout": {}}
	if err := rejectUnknown(values, allowed, label); err != nil {
		return ExecutionOptions{}, err
	}

	return decodeExecutionValues(values, label)
}

func decodeExecutionValues(values map[string]runtime.Value, label string) (ExecutionOptions, error) {
	var options ExecutionOptions

	if temperature, found, err := numberOption(values, "temperature", label); err != nil {
		return options, err
	} else if found {
		if temperature < 0 || temperature > 2 {
			return options, NewError(ErrInvalidOptions, label+".temperature must be between 0 and 2")
		}
		options.Temperature = &temperature
	}

	if maxTokens, found, err := intOption(values, "maxOutputTokens", label); err != nil {
		return options, err
	} else if found {
		if maxTokens <= 0 {
			return options, NewError(ErrInvalidOptions, label+".maxOutputTokens must be positive")
		}
		options.MaxOutputTokens = maxTokens
	}

	if timeout, found, err := intOption(values, "timeout", label); err != nil {
		return options, err
	} else if found {
		if timeout < 0 {
			return options, NewError(ErrInvalidOptions, label+".timeout must be nonnegative")
		}
		if timeout > math.MaxInt64/int64(time.Millisecond) {
			return options, NewError(ErrInvalidOptions, label+".timeout is too large")
		}
		options.Timeout = time.Duration(timeout) * time.Millisecond
	}

	return options, nil
}
