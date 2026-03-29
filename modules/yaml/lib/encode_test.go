package lib

import (
	"context"
	"testing"

	"github.com/MontFerret/contrib/modules/yaml/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestEncodeLib(t *testing.T) {
	ctx := context.Background()

	t.Run("encodes runtime value to string", func(t *testing.T) {
		value := runtime.NewObjectWith(map[string]runtime.Value{
			"name": runtime.NewString("Alice"),
			"age":  runtime.NewInt(30),
		})

		result, err := Encode(ctx, value)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		text, ok := result.(runtime.String)
		if !ok {
			t.Fatalf("expected runtime.String, got %T", result)
		}

		decoded, err := core.Decode(ctx, text)
		if err != nil {
			t.Fatalf("unexpected decode error: %v", err)
		}

		if decoded.String() != value.String() {
			t.Fatalf("round-trip mismatch: got %s want %s", decoded.String(), value.String())
		}
	})

	t.Run("rejects unsupported value type", func(t *testing.T) {
		_, err := Encode(ctx, runtime.NewBinary([]byte("yaml")))
		if err == nil {
			t.Fatal("expected unsupported type error")
		}
	})

	t.Run("validates arity", func(t *testing.T) {
		if _, err := Encode(ctx); err == nil {
			t.Fatal("expected error for missing arg")
		}

		if _, err := Encode(ctx, runtime.NewString("a"), runtime.NewString("b")); err == nil {
			t.Fatal("expected error for extra arg")
		}
	})
}
