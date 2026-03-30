package core

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func assertStageError(t *testing.T, err error, stage Stage, url string) *Error {
	t.Helper()

	var target *Error
	if !errors.As(err, &target) {
		t.Fatalf("expected *Error, got %T (%v)", err, err)
	}

	if target.Stage != stage {
		t.Fatalf("expected stage %q, got %q", stage, target.Stage)
	}

	if url != "" && !strings.Contains(target.URL, url) {
		t.Fatalf("expected error URL %q to contain %q", target.URL, url)
	}

	return target
}

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
