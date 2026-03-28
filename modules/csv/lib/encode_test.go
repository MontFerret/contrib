package lib

import (
	"context"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestEncodeLib(t *testing.T) {
	ctx := context.Background()

	t.Run("too few args", func(t *testing.T) {
		_, err := Encode(ctx)
		if err == nil {
			t.Fatal("expected error for no args")
		}
	})

	t.Run("too many args", func(t *testing.T) {
		_, err := Encode(ctx, runtime.NewArray(0), runtime.NewObject(), runtime.NewString("extra"))
		if err == nil {
			t.Fatal("expected error for too many args")
		}
	})

	t.Run("wrong options type", func(t *testing.T) {
		_, err := Encode(ctx, runtime.NewArray(0), runtime.NewString("not-a-map"))
		if err == nil {
			t.Fatal("expected error for wrong options type")
		}
	})

	t.Run("wrong arg type", func(t *testing.T) {
		_, err := Encode(ctx, runtime.NewString("not-an-array"))
		if err == nil {
			t.Fatal("expected error for wrong arg type")
		}
	})

	t.Run("invalid option payload returns error", func(t *testing.T) {
		opts := runtime.NewObjectWith(map[string]runtime.Value{
			"delimiter": runtime.NewInt(1),
		})

		_, err := Encode(ctx, runtime.NewArray(0), opts)
		if err == nil {
			t.Fatal("expected error for invalid option payload")
		}
	})

	t.Run("successful encode path", func(t *testing.T) {
		opts := runtime.NewObjectWith(map[string]runtime.Value{
			"delimiter": runtime.NewString(";"),
		})

		data := runtime.NewArrayWith(
			runtime.NewArrayWith(runtime.NewString("name"), runtime.NewString("age")),
			runtime.NewArrayWith(runtime.NewString("Alice"), runtime.NewInt(30)),
		)

		result, err := Encode(ctx, data, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.String() != "name;age\nAlice;30\n" {
			t.Fatalf("unexpected encode result: %q", result.String())
		}
	})
}
