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

func mustObjectField(t *testing.T, ctx context.Context, obj *runtime.Object, key string) runtime.Value {
	t.Helper()

	value, err := obj.Get(ctx, runtime.NewString(key))
	if err != nil {
		t.Fatalf("unexpected error getting %q: %v", key, err)
	}

	return value
}
