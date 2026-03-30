package core

import (
	"context"
	"strings"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestEncodeCore(t *testing.T) {
	ctx := context.Background()

	t.Run("encodes scalar null", func(t *testing.T) {
		result, err := Encode(ctx, runtime.None)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if strings.TrimSpace(result) != "null" {
			t.Fatalf("unexpected encode result: %q", result)
		}
	})

	t.Run("encodes arrays and objects", func(t *testing.T) {
		value := runtime.NewObjectWith(map[string]runtime.Value{
			"name":   runtime.NewString("Alice"),
			"age":    runtime.NewInt(30),
			"active": runtime.True,
			"tags": runtime.NewArrayWith(
				runtime.NewString("admin"),
				runtime.NewString("ops"),
			),
		})

		result, err := Encode(ctx, value)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		decoded, err := Decode(ctx, runtime.NewString(result))
		if err != nil {
			t.Fatalf("unexpected decode error: %v", err)
		}

		if decoded.String() != value.String() {
			t.Fatalf("round-trip mismatch: got %s want %s", decoded.String(), value.String())
		}
	})

	t.Run("round trips representative document", func(t *testing.T) {
		input := `
name: Alice
age: 30
features:
  - yaml
  - ferret
meta:
  enabled: true
  score: 4.5
`

		decoded, err := Decode(ctx, runtime.NewString(input))
		if err != nil {
			t.Fatalf("unexpected decode error: %v", err)
		}

		encoded, err := Encode(ctx, decoded)
		if err != nil {
			t.Fatalf("unexpected encode error: %v", err)
		}

		decodedAgain, err := Decode(ctx, runtime.NewString(encoded))
		if err != nil {
			t.Fatalf("unexpected round-trip decode error: %v", err)
		}

		if decoded.String() != decodedAgain.String() {
			t.Fatalf("round-trip mismatch: got %s want %s", decodedAgain.String(), decoded.String())
		}
	})

	t.Run("rejects unsupported binary value", func(t *testing.T) {
		_, err := Encode(ctx, runtime.NewBinary([]byte("yaml")))
		if err == nil {
			t.Fatal("expected unsupported type error")
		}

		if _, ok := err.(*Error); !ok {
			t.Fatalf("expected *YAMLError, got %T", err)
		}

		if !strings.Contains(err.Error(), "unsupported value type for YAML encoding") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
