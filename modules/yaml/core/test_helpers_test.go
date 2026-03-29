package core

import (
	"context"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func mustRuntimeObject(t *testing.T, val runtime.Value) *runtime.Object {
	t.Helper()

	obj, ok := val.(*runtime.Object)
	if !ok {
		t.Fatalf("expected *runtime.Object, got %T", val)
	}

	return obj
}

func mustRuntimeArray(t *testing.T, val runtime.Value) *runtime.Array {
	t.Helper()

	arr, ok := val.(*runtime.Array)
	if !ok {
		t.Fatalf("expected *runtime.Array, got %T", val)
	}

	return arr
}

func mustArrayAtIndex(t *testing.T, ctx context.Context, arr *runtime.Array, idx int) runtime.Value {
	t.Helper()

	value, err := arr.At(ctx, runtime.Int(idx))
	if err != nil {
		t.Fatalf("unexpected error at index %d: %v", idx, err)
	}

	return value
}

func mustObjectField(t *testing.T, ctx context.Context, obj *runtime.Object, key string) runtime.Value {
	t.Helper()

	value, err := obj.Get(ctx, runtime.NewString(key))
	if err != nil {
		t.Fatalf("unexpected error getting %q: %v", key, err)
	}

	if value == runtime.None {
		t.Fatalf("expected field %q to exist", key)
	}

	return value
}

func mustObjectFieldObject(t *testing.T, ctx context.Context, obj *runtime.Object, key string) *runtime.Object {
	t.Helper()

	return mustRuntimeObject(t, mustObjectField(t, ctx, obj, key))
}

func mustObjectFieldArray(t *testing.T, ctx context.Context, obj *runtime.Object, key string) *runtime.Array {
	t.Helper()

	return mustRuntimeArray(t, mustObjectField(t, ctx, obj, key))
}

func assertArrayLen(t *testing.T, ctx context.Context, arr *runtime.Array, expected int) {
	t.Helper()

	length, err := arr.Length(ctx)
	if err != nil {
		t.Fatalf("unexpected length error: %v", err)
	}

	if int(length) != expected {
		t.Fatalf("expected length %d, got %d", expected, int(length))
	}
}

func assertRuntimeStringValue(t *testing.T, value runtime.Value, expected string) {
	t.Helper()

	str, ok := value.(runtime.String)
	if !ok {
		t.Fatalf("expected runtime.String, got %T", value)
	}

	if str.String() != expected {
		t.Fatalf("expected %q, got %q", expected, str.String())
	}
}

func assertRuntimeIntValue(t *testing.T, value runtime.Value, expected int64) {
	t.Helper()

	i, ok := value.(runtime.Int)
	if !ok {
		t.Fatalf("expected runtime.Int, got %T", value)
	}

	if int64(i) != expected {
		t.Fatalf("expected %d, got %d", expected, int64(i))
	}
}

func assertRuntimeFloatValue(t *testing.T, value runtime.Value, expected float64) {
	t.Helper()

	f, ok := value.(runtime.Float)
	if !ok {
		t.Fatalf("expected runtime.Float, got %T", value)
	}

	if float64(f) != expected {
		t.Fatalf("expected %f, got %f", expected, float64(f))
	}
}

func assertRuntimeBoolValue(t *testing.T, value runtime.Value, expected bool) {
	t.Helper()

	b, ok := value.(runtime.Boolean)
	if !ok {
		t.Fatalf("expected runtime.Boolean, got %T", value)
	}

	if bool(b) != expected {
		t.Fatalf("expected %t, got %t", expected, bool(b))
	}
}

func assertObjectFieldString(t *testing.T, ctx context.Context, obj *runtime.Object, key, expected string) {
	t.Helper()

	assertRuntimeStringValue(t, mustObjectField(t, ctx, obj, key), expected)
}

func assertObjectFieldInt(t *testing.T, ctx context.Context, obj *runtime.Object, key string, expected int64) {
	t.Helper()

	assertRuntimeIntValue(t, mustObjectField(t, ctx, obj, key), expected)
}

func assertObjectFieldFloat(t *testing.T, ctx context.Context, obj *runtime.Object, key string, expected float64) {
	t.Helper()

	assertRuntimeFloatValue(t, mustObjectField(t, ctx, obj, key), expected)
}

func assertObjectFieldBool(t *testing.T, ctx context.Context, obj *runtime.Object, key string, expected bool) {
	t.Helper()

	assertRuntimeBoolValue(t, mustObjectField(t, ctx, obj, key), expected)
}

func assertObjectFieldNone(t *testing.T, ctx context.Context, obj *runtime.Object, key string) {
	t.Helper()

	value, err := obj.Get(ctx, runtime.NewString(key))
	if err != nil {
		t.Fatalf("unexpected error getting %q: %v", key, err)
	}

	if value != runtime.None {
		t.Fatalf("expected None for %q, got %v", key, value)
	}
}
