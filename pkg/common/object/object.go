package object

import (
	"context"
	"fmt"
	"time"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// RequireMap returns value as a map or an owner-qualified object error.
func RequireMap(value runtime.Value, owner string) (runtime.Map, error) {
	if runtime.TypeNone.Is(value) {
		return nil, fmt.Errorf("%s must be an object", owner)
	}

	obj, ok := value.(runtime.Map)
	if !ok {
		return nil, fmt.Errorf("%s must be an object", owner)
	}

	return obj, nil
}

// Value looks up a map field by string key.
func Value(ctx context.Context, obj runtime.Map, key string) (runtime.Value, bool, error) {
	value, found, err := obj.Lookup(ctx, runtime.NewString(key))
	if err != nil {
		return runtime.None, false, err
	}

	return value, found, nil
}

// ValueAny looks up a map field by key or aliases.
func ValueAny(ctx context.Context, obj runtime.Map, key string, aliases ...string) (runtime.Value, bool, error) {
	keys := append([]string{key}, aliases...)

	for _, item := range keys {
		value, found, err := Value(ctx, obj, item)
		if err != nil {
			return runtime.None, false, err
		}
		if found && value != runtime.None {
			return value, true, nil
		}
	}

	return runtime.None, false, nil
}

// String looks up an optional string field and formats type errors as owner.key.
func String(ctx context.Context, obj runtime.Map, key, owner string) (string, bool, error) {
	value, found, err := Value(ctx, obj, key)
	if err != nil || !found {
		return "", found, err
	}

	str, ok := value.(runtime.String)
	if !ok {
		return "", true, fmt.Errorf("%s.%s must be a string", owner, key)
	}

	return str.String(), true, nil
}

// MillisDuration looks up an optional non-negative integer duration in milliseconds.
func MillisDuration(ctx context.Context, obj runtime.Map, key, owner string) (time.Duration, bool, error) {
	value, found, err := Value(ctx, obj, key)
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

// AliasString looks up an optional string field by key or aliases.
func AliasString(ctx context.Context, obj runtime.Map, key string, aliases ...string) (string, bool, error) {
	value, found, err := ValueAny(ctx, obj, key, aliases...)
	if err != nil || !found {
		return "", found, err
	}

	if err := runtime.ValidateType(value, runtime.TypeString); err != nil {
		return "", true, err
	}

	return value.String(), true, nil
}

// AliasBool looks up an optional boolean field by key or aliases.
func AliasBool(ctx context.Context, obj runtime.Map, key string, aliases ...string) (bool, bool, error) {
	value, found, err := ValueAny(ctx, obj, key, aliases...)
	if err != nil || !found {
		return false, found, err
	}

	if err := runtime.ValidateType(value, runtime.TypeBoolean); err != nil {
		return false, true, err
	}

	return bool(value.(runtime.Boolean)), true, nil
}

// AliasInt looks up an optional integer field by key or aliases.
func AliasInt(ctx context.Context, obj runtime.Map, key string, aliases ...string) (runtime.Int, bool, error) {
	value, found, err := ValueAny(ctx, obj, key, aliases...)
	if err != nil || !found {
		return 0, found, err
	}

	if err := runtime.ValidateType(value, runtime.TypeInt); err != nil {
		return 0, true, err
	}

	return value.(runtime.Int), true, nil
}

// StringMap converts a Ferret map into a string map with owner-qualified errors.
func StringMap(ctx context.Context, input runtime.Map, owner string) (map[string]string, error) {
	out := make(map[string]string)

	if err := input.ForEach(ctx, func(_ context.Context, value, key runtime.Value) (runtime.Boolean, error) {
		name, ok := key.(runtime.String)
		if !ok {
			return false, fmt.Errorf("%s keys must be strings", owner)
		}

		text, ok := value.(runtime.String)
		if !ok {
			return false, fmt.Errorf("%s values must be strings", owner)
		}

		out[name.String()] = text.String()

		return true, nil
	}); err != nil {
		return nil, err
	}

	return out, nil
}
