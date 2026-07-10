package core

import (
	"context"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func testModel(executor *fakeExecutor) Model {
	return NewStatelessModel("test", "opaque-model", executor, executor)
}

func object(values map[string]runtime.Value) *runtime.Object {
	return runtime.NewObjectWith(values)
}

func requireCode(t *testing.T, err error, expected ErrorCode) {
	t.Helper()

	actual, ok := CodeOf(err)
	if !ok {
		t.Fatalf("expected typed error %s, got %v", expected, err)
	}
	if actual != expected {
		t.Fatalf("expected error code %s, got %s (%v)", expected, actual, err)
	}
}

func objectValue(t *testing.T, value runtime.Value, key string) runtime.Value {
	t.Helper()

	obj, ok := value.(runtime.Map)
	if !ok {
		t.Fatalf("expected object, got %T", value)
	}

	result, err := obj.Get(context.Background(), runtime.NewString(key))
	if err != nil {
		t.Fatal(err)
	}

	return result
}
