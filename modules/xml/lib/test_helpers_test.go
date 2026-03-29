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

	val, err := obj.Get(ctx, runtime.NewString(key))
	if err != nil {
		t.Fatalf("unexpected error getting %q: %v", key, err)
	}

	if val == runtime.None {
		t.Fatalf("expected field %q to exist", key)
	}

	return val
}

func mustObjectFieldObject(t *testing.T, ctx context.Context, obj *runtime.Object, key string) *runtime.Object {
	t.Helper()

	return mustRuntimeObject(t, mustObjectField(t, ctx, obj, key))
}

func mustObjectFieldArray(t *testing.T, ctx context.Context, obj *runtime.Object, key string) *runtime.Array {
	t.Helper()

	return mustRuntimeArray(t, mustObjectField(t, ctx, obj, key))
}

func mustArrayAtIndex(t *testing.T, ctx context.Context, arr *runtime.Array, idx int) runtime.Value {
	t.Helper()

	val, err := arr.At(ctx, runtime.Int(idx))
	if err != nil {
		t.Fatalf("unexpected error at index %d: %v", idx, err)
	}

	return val
}

func mustArrayObjectAtIndex(t *testing.T, ctx context.Context, arr *runtime.Array, idx int) *runtime.Object {
	t.Helper()

	return mustRuntimeObject(t, mustArrayAtIndex(t, ctx, arr, idx))
}

func assertRuntimeStringValue(t *testing.T, val runtime.Value, expected string) {
	t.Helper()

	str, ok := val.(runtime.String)
	if !ok {
		t.Fatalf("expected runtime.String, got %T", val)
	}

	if str.String() != expected {
		t.Fatalf("expected %q, got %q", expected, str.String())
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

func assertObjectFieldString(t *testing.T, ctx context.Context, obj *runtime.Object, key, expected string) {
	t.Helper()

	assertRuntimeStringValue(t, mustObjectField(t, ctx, obj, key), expected)
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

func xmlDocument(root runtime.Value) *runtime.Object {
	return runtime.NewObjectWith(map[string]runtime.Value{
		"type": runtime.NewString("document"),
		"root": root,
	})
}

func xmlElement(name string, attrs map[string]runtime.Value, children ...runtime.Value) *runtime.Object {
	attrMap := runtime.NewObject()
	for key, value := range attrs {
		_ = attrMap.Set(context.Background(), runtime.NewString(key), value)
	}

	childArray := runtime.NewArray(0)
	for _, child := range children {
		_ = childArray.Append(context.Background(), child)
	}

	return runtime.NewObjectWith(map[string]runtime.Value{
		"type":     runtime.NewString("element"),
		"name":     runtime.NewString(name),
		"attrs":    attrMap,
		"children": childArray,
	})
}

func xmlText(value string) *runtime.Object {
	return runtime.NewObjectWith(map[string]runtime.Value{
		"type":  runtime.NewString("text"),
		"value": runtime.NewString(value),
	})
}
