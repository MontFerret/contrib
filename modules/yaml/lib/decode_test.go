package lib

import (
	"context"
	"testing"

	"github.com/MontFerret/contrib/modules/yaml/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestDecodeLib(t *testing.T) {
	ctx := context.Background()

	t.Run("decodes string input", func(t *testing.T) {
		result, err := Decode(ctx, runtime.NewString("name: Alice\nage: 30\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		obj := mustRuntimeObject(t, result)
		assertObjectFieldString(t, ctx, obj, "name", "Alice")
		assertObjectFieldInt(t, ctx, obj, "age", 30)
	})

	t.Run("accepts binary input", func(t *testing.T) {
		result, err := Decode(ctx, runtime.NewBinary([]byte("active: true\n")))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		obj := mustRuntimeObject(t, result)
		assertObjectFieldBool(t, ctx, obj, "active", true)
	})

	t.Run("rejects non text input", func(t *testing.T) {
		_, err := Decode(ctx, runtime.NewInt(42))
		if err == nil {
			t.Fatal("expected non-text input error")
		}
	})

	t.Run("rejects malformed yaml", func(t *testing.T) {
		_, err := Decode(ctx, runtime.NewString("name: [unterminated"))
		if err == nil {
			t.Fatal("expected malformed YAML error")
		}

		if _, ok := err.(*core.Error); !ok {
			t.Fatalf("expected *core.YAMLError, got %T", err)
		}
	})
}

func TestDecodeAllLib(t *testing.T) {
	ctx := context.Background()

	t.Run("returns runtime array", func(t *testing.T) {
		result, err := DecodeAll(ctx, runtime.NewString("---\na: 1\n---\ntrue\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		arr := mustRuntimeArray(t, result)
		assertArrayLen(t, ctx, arr, 2)
	})

	t.Run("validates arity", func(t *testing.T) {
		if _, err := DecodeAll(ctx); err == nil {
			t.Fatal("expected error for missing arg")
		}

		if _, err := DecodeAll(ctx, runtime.NewString("a: 1"), runtime.NewString("extra")); err == nil {
			t.Fatal("expected error for extra arg")
		}
	})
}
