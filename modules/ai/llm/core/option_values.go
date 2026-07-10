package core

import (
	"context"
	"fmt"
	"math"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func optionValues(ctx context.Context, value runtime.Value, label string) (map[string]runtime.Value, error) {
	if value == nil || runtime.TypeNone.Is(value) {
		return map[string]runtime.Value{}, nil
	}

	object, ok := value.(runtime.Map)
	if !ok {
		return nil, NewError(ErrInvalidOptions, label+" must be an object")
	}

	values := make(map[string]runtime.Value)
	err := object.ForEach(ctx, func(_ context.Context, value, key runtime.Value) (runtime.Boolean, error) {
		name, ok := key.(runtime.String)
		if !ok {
			return runtime.False, NewError(ErrInvalidOptions, label+" keys must be strings")
		}

		values[name.String()] = value

		return runtime.True, nil
	})
	if err != nil {
		return nil, err
	}

	return values, nil
}

func stringOption(values map[string]runtime.Value, key, label string) (string, bool, error) {
	value, found := values[key]
	if !found {
		return "", false, nil
	}

	text, ok := value.(runtime.String)
	if !ok {
		return "", false, NewError(ErrInvalidOptions, fmt.Sprintf("%s.%s must be a string", label, key))
	}

	return text.String(), true, nil
}

func boolOption(values map[string]runtime.Value, key, label string) (bool, bool, error) {
	value, found := values[key]
	if !found {
		return false, false, nil
	}

	flag, ok := value.(runtime.Boolean)
	if !ok {
		return false, false, NewError(ErrInvalidOptions, fmt.Sprintf("%s.%s must be a boolean", label, key))
	}

	return bool(flag), true, nil
}

func intOption(values map[string]runtime.Value, key, label string) (int64, bool, error) {
	value, found := values[key]
	if !found {
		return 0, false, nil
	}

	integer, ok := value.(runtime.Int)
	if !ok {
		return 0, false, NewError(ErrInvalidOptions, fmt.Sprintf("%s.%s must be an integer", label, key))
	}

	return int64(integer), true, nil
}

func numberOption(values map[string]runtime.Value, key, label string) (float64, bool, error) {
	value, found := values[key]
	if !found {
		return 0, false, nil
	}

	var number float64
	switch typed := value.(type) {
	case runtime.Int:
		number = float64(typed)
	case runtime.Float:
		number = float64(typed)
	default:
		return 0, false, NewError(ErrInvalidOptions, fmt.Sprintf("%s.%s must be a number", label, key))
	}

	if math.IsNaN(number) || math.IsInf(number, 0) {
		return 0, false, NewError(ErrInvalidOptions, fmt.Sprintf("%s.%s must be finite", label, key))
	}

	return number, true, nil
}

func rejectUnknown(values map[string]runtime.Value, allowed map[string]struct{}, label string) error {
	for key := range values {
		if _, ok := allowed[key]; !ok {
			return NewError(ErrInvalidOptions, fmt.Sprintf("unknown %s key %q", label, key))
		}
	}

	return nil
}
