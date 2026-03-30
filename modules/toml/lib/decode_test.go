package lib

import (
	"context"
	"testing"

	"github.com/MontFerret/contrib/modules/toml/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestDecodeLib(t *testing.T) {
	ctx := context.Background()

	t.Run("decodes string input", func(t *testing.T) {
		result, err := Decode(ctx, runtime.NewString("title = \"Ferret\"\n[server]\nport = 8080\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		obj := mustRuntimeObject(t, result)
		assertRuntimeStringValue(t, mustObjectField(t, ctx, obj, "title"), "Ferret")

		server := mustRuntimeObject(t, mustObjectField(t, ctx, obj, "server"))
		assertRuntimeIntValue(t, mustObjectField(t, ctx, server, "port"), 8080)
	})

	t.Run("accepts binary input", func(t *testing.T) {
		result, err := Decode(ctx, runtime.NewBinary([]byte("active = true\n")))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		obj := mustRuntimeObject(t, result)
		assertRuntimeBoolValue(t, mustObjectField(t, ctx, obj, "active"), true)
	})

	t.Run("accepts native datetime option", func(t *testing.T) {
		options := runtime.NewObjectWith(map[string]runtime.Value{
			"datetime": runtime.NewString(core.DecodeDateTimeNative),
		})

		result, err := Decode(ctx, runtime.NewString("released = 1979-05-27T07:32:00\n"), options)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		obj := mustRuntimeObject(t, result)
		if _, ok := mustObjectField(t, ctx, obj, "released").(runtime.DateTime); !ok {
			t.Fatalf("expected runtime.DateTime, got %T", mustObjectField(t, ctx, obj, "released"))
		}
	})

	t.Run("rejects strict false until non-strict mode exists", func(t *testing.T) {
		options := runtime.NewObjectWith(map[string]runtime.Value{
			"strict": runtime.False,
		})

		_, err := Decode(ctx, runtime.NewString("title = \"Ferret\"\n"), options)
		if err == nil {
			t.Fatal("expected strict=false error")
		}

		if _, ok := err.(*core.TOMLError); !ok {
			t.Fatalf("expected *core.TOMLError, got %T", err)
		}
	})

	t.Run("rejects malformed toml", func(t *testing.T) {
		_, err := Decode(ctx, runtime.NewString("title = [unterminated"))
		if err == nil {
			t.Fatal("expected malformed TOML error")
		}

		if _, ok := err.(*core.TOMLError); !ok {
			t.Fatalf("expected *core.TOMLError, got %T", err)
		}
	})

	t.Run("validates arity", func(t *testing.T) {
		if _, err := Decode(ctx); err == nil {
			t.Fatal("expected error for missing arg")
		}

		if _, err := Decode(ctx, runtime.NewString("a = 1"), runtime.NewObject(), runtime.NewString("extra")); err == nil {
			t.Fatal("expected error for extra arg")
		}
	})
}
