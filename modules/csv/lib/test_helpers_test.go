package lib

import (
	"context"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func mustIterate(t *testing.T, ctx context.Context, val runtime.Value) runtime.Iterator {
	t.Helper()

	iterable, ok := val.(runtime.Iterable)
	if !ok {
		t.Fatalf("expected runtime.Iterable, got %T", val)
	}

	iter, err := iterable.Iterate(ctx)
	if err != nil {
		t.Fatalf("unexpected iterate error: %v", err)
	}

	return iter
}

func mustRuntimeArray(t *testing.T, val runtime.Value) *runtime.Array {
	t.Helper()

	arr, ok := val.(*runtime.Array)
	if !ok {
		t.Fatalf("expected *runtime.Array, got %T", val)
	}

	return arr
}

func mustRuntimeObject(t *testing.T, val runtime.Value) *runtime.Object {
	t.Helper()

	obj, ok := val.(*runtime.Object)
	if !ok {
		t.Fatalf("expected *runtime.Object, got %T", val)
	}

	return obj
}

func mustObjectAtIndex(t *testing.T, ctx context.Context, arr *runtime.Array, idx int) *runtime.Object {
	t.Helper()

	val, err := arr.At(ctx, runtime.Int(idx))
	if err != nil {
		t.Fatalf("unexpected error at index %d: %v", idx, err)
	}

	return mustRuntimeObject(t, val)
}

func mustArrayAtIndex(t *testing.T, ctx context.Context, arr *runtime.Array, idx int) *runtime.Array {
	t.Helper()

	val, err := arr.At(ctx, runtime.Int(idx))
	if err != nil {
		t.Fatalf("unexpected error at index %d: %v", idx, err)
	}

	return mustRuntimeArray(t, val)
}

func assertRuntimeArrayLen(t *testing.T, ctx context.Context, arr *runtime.Array, expected int) {
	t.Helper()

	length, err := arr.Length(ctx)
	if err != nil {
		t.Fatalf("unexpected length error: %v", err)
	}

	if int(length) != expected {
		t.Fatalf("expected length %d, got %d", expected, int(length))
	}
}

func assertRuntimeIntValue(t *testing.T, val runtime.Value, expected int64) {
	t.Helper()

	i, ok := val.(runtime.Int)
	if !ok {
		t.Fatalf("expected runtime.Int, got %T", val)
	}

	if int64(i) != expected {
		t.Fatalf("expected %d, got %d", expected, int64(i))
	}
}

func assertObjectField(t *testing.T, ctx context.Context, obj *runtime.Object, key, expected string) {
	t.Helper()

	val, err := obj.Get(ctx, runtime.NewString(key))
	if err != nil {
		t.Fatalf("unexpected error getting %q: %v", key, err)
	}

	if val.String() != expected {
		t.Fatalf("field %q: expected %q, got %q", key, expected, val.String())
	}
}
