package lib

import (
	"context"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestDecodeLib(t *testing.T) {
	ctx := context.Background()

	t.Run("too few args", func(t *testing.T) {
		_, err := Decode(ctx)
		if err == nil {
			t.Fatal("expected error for no args")
		}
	})

	t.Run("too many args", func(t *testing.T) {
		_, err := Decode(ctx, runtime.NewString("a"), runtime.NewObject(), runtime.NewString("extra"))
		if err == nil {
			t.Fatal("expected error for too many args")
		}
	})

	t.Run("wrong arg type", func(t *testing.T) {
		_, err := Decode(ctx, runtime.NewInt(42))
		if err == nil {
			t.Fatal("expected error for wrong arg type")
		}
	})

	t.Run("valid call string only", func(t *testing.T) {
		result, err := Decode(ctx, runtime.NewString("name,age\nAlice,30"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		arr, ok := result.(*runtime.Array)
		if !ok {
			t.Fatalf("expected *runtime.Array, got %T", result)
		}

		length, _ := arr.Length(ctx)
		if int(length) != 1 {
			t.Fatalf("expected 1 row, got %d", int(length))
		}
	})

	t.Run("valid call with options", func(t *testing.T) {
		opts := runtime.NewObjectWith(map[string]runtime.Value{
			"delimiter": runtime.NewString(";"),
		})

		result, err := Decode(ctx, runtime.NewString("name;age\nAlice;30"), opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		arr, ok := result.(*runtime.Array)
		if !ok {
			t.Fatalf("expected *runtime.Array, got %T", result)
		}

		length, _ := arr.Length(ctx)
		if int(length) != 1 {
			t.Fatalf("expected 1 row, got %d", int(length))
		}
	})

	t.Run("wrong options type", func(t *testing.T) {
		_, err := Decode(ctx, runtime.NewString("a,b"), runtime.NewString("not-a-map"))
		if err == nil {
			t.Fatal("expected error for wrong options type")
		}
	})
}

func TestDecodeRowsLib(t *testing.T) {
	ctx := context.Background()

	t.Run("basic call", func(t *testing.T) {
		result, err := DecodeRows(ctx, runtime.NewString("a,b\nc,d"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		arr, ok := result.(*runtime.Array)
		if !ok {
			t.Fatalf("expected *runtime.Array, got %T", result)
		}

		length, _ := arr.Length(ctx)
		if int(length) != 2 {
			t.Fatalf("expected 2 rows, got %d", int(length))
		}
	})
}
