package core

import (
	"context"
	"math"
	"time"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type executionOptionsInput struct {
	Temperature     *float64 `json:"temperature"`
	MaxOutputTokens *int64   `json:"maxOutputTokens"`
	Timeout         *int64   `json:"timeout"`
}

// DecodeExecutionOptions validates common provider execution policy.
func DecodeExecutionOptions(ctx context.Context, value runtime.Value) (ExecutionOptions, error) {
	const label = "execution options"
	input, err := decodeOptionObject[executionOptionsInput](ctx, value, label)
	if err != nil {
		return ExecutionOptions{}, err
	}

	return decodeExecutionInput(input, label)
}

func decodeExecutionInput(input executionOptionsInput, label string) (ExecutionOptions, error) {
	var options ExecutionOptions

	if input.Temperature != nil {
		if math.IsNaN(*input.Temperature) || math.IsInf(*input.Temperature, 0) {
			return options, NewError(ErrInvalidOptions, label+".temperature must be finite")
		}
		if *input.Temperature < 0 || *input.Temperature > 2 {
			return options, NewError(ErrInvalidOptions, label+".temperature must be between 0 and 2")
		}

		temperature := *input.Temperature
		options.Temperature = &temperature
	}

	if input.MaxOutputTokens != nil {
		if *input.MaxOutputTokens <= 0 {
			return options, NewError(ErrInvalidOptions, label+".maxOutputTokens must be positive")
		}

		options.MaxOutputTokens = *input.MaxOutputTokens
	}

	if input.Timeout != nil {
		if *input.Timeout < 0 {
			return options, NewError(ErrInvalidOptions, label+".timeout must be nonnegative")
		}

		if *input.Timeout > math.MaxInt64/int64(time.Millisecond) {
			return options, NewError(ErrInvalidOptions, label+".timeout is too large")
		}

		options.Timeout = time.Duration(*input.Timeout) * time.Millisecond
	}

	return options, nil
}
