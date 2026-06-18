package core

import (
	"context"
	"fmt"
	"time"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func requireMap(_ context.Context, value runtime.Value, owner string) (runtime.Map, error) {
	if runtime.TypeNone.Is(value) {
		return nil, fmt.Errorf("%s must be an object", owner)
	}

	obj, ok := value.(runtime.Map)
	if !ok {
		return nil, fmt.Errorf("%s must be an object", owner)
	}

	return obj, nil
}

func lookupValue(ctx context.Context, obj runtime.Map, key string) (runtime.Value, bool, error) {
	value, found, err := obj.Lookup(ctx, runtime.NewString(key))
	if err != nil {
		return runtime.None, false, err
	}

	return value, found, nil
}

func lookupString(ctx context.Context, obj runtime.Map, key, owner string) (string, bool, error) {
	value, found, err := lookupValue(ctx, obj, key)
	if err != nil || !found {
		return "", found, err
	}

	str, ok := value.(runtime.String)
	if !ok {
		return "", true, fmt.Errorf("%s.%s must be a string", owner, key)
	}

	return str.String(), true, nil
}

func lookupDuration(ctx context.Context, obj runtime.Map, key, owner string) (time.Duration, bool, error) {
	value, found, err := lookupValue(ctx, obj, key)
	if err != nil || !found {
		return 0, found, err
	}

	timeout, err := runtime.ToInt(ctx, value)
	if err != nil {
		return 0, true, fmt.Errorf("%s.%s must be an integer number of milliseconds: %w", owner, key, err)
	}
	if timeout < 0 {
		return 0, true, fmt.Errorf("%s.%s must be greater than or equal to 0", owner, key)
	}

	return time.Duration(timeout) * time.Millisecond, true, nil
}
