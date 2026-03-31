package lib

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

func mustObjectField(t *testing.T, ctx context.Context, obj *runtime.Object, key string) runtime.Value {
	t.Helper()

	value, err := obj.Get(ctx, runtime.NewString(key))
	if err != nil {
		t.Fatalf("unexpected error getting %q: %v", key, err)
	}

	return value
}

func mustRuntimeArray(t *testing.T, val runtime.Value) *runtime.Array {
	t.Helper()

	arr, ok := val.(*runtime.Array)
	if !ok {
		t.Fatalf("expected *runtime.Array, got %T", val)
	}

	return arr
}

func mustArrayStrings(t *testing.T, ctx context.Context, arr *runtime.Array) []string {
	t.Helper()

	length, err := arr.Length(ctx)
	if err != nil {
		t.Fatalf("unexpected array length error: %v", err)
	}

	result := make([]string, 0, int(length))
	for idx := runtime.Int(0); idx < length; idx++ {
		value, err := arr.At(ctx, idx)
		if err != nil {
			t.Fatalf("unexpected array get error: %v", err)
		}

		text, ok := value.(runtime.String)
		if !ok {
			t.Fatalf("expected runtime.String, got %T", value)
		}

		result = append(result, text.String())
	}

	return result
}
