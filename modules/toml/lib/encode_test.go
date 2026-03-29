package lib

import (
	"context"
	"strings"
	"testing"

	"github.com/MontFerret/contrib/modules/toml/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestEncodeLib(t *testing.T) {
	ctx := context.Background()

	t.Run("encodes runtime value to string", func(t *testing.T) {
		value := runtime.NewObjectWith(map[string]runtime.Value{
			"title": runtime.NewString("Ferret"),
			"server": runtime.NewObjectWith(map[string]runtime.Value{
				"port": runtime.NewInt(8080),
			}),
		})

		result, err := Encode(ctx, value)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		text, ok := result.(runtime.String)
		if !ok {
			t.Fatalf("expected runtime.String, got %T", result)
		}

		decoded, err := core.Decode(ctx, text, core.DefaultDecodeOptions())
		if err != nil {
			t.Fatalf("unexpected decode error: %v", err)
		}

		if decoded.String() != value.String() {
			t.Fatalf("round-trip mismatch: got %s want %s", decoded.String(), value.String())
		}
	})

	t.Run("applies encode options", func(t *testing.T) {
		value := runtime.NewObjectWith(map[string]runtime.Value{
			"b": runtime.NewInt(2),
			"a": runtime.NewInt(1),
		})

		options := runtime.NewObjectWith(map[string]runtime.Value{
			"sort_keys": runtime.True,
		})

		result, err := Encode(ctx, value, options)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		text := result.(runtime.String).String()
		if !strings.HasPrefix(text, "a = 1\nb = 2") {
			t.Fatalf("expected sorted output, got %q", text)
		}
	})

	t.Run("rejects unsupported value type", func(t *testing.T) {
		_, err := Encode(ctx, runtime.NewBinary([]byte("toml")))
		if err == nil {
			t.Fatal("expected unsupported type error")
		}
	})

	t.Run("validates arity", func(t *testing.T) {
		if _, err := Encode(ctx); err == nil {
			t.Fatal("expected error for missing arg")
		}

		if _, err := Encode(ctx, runtime.NewObject(), runtime.NewObject(), runtime.NewString("extra")); err == nil {
			t.Fatal("expected error for extra arg")
		}
	})
}
