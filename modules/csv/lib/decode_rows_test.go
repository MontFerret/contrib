package lib

import (
	"context"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestDecodeRowsLib(t *testing.T) {
	ctx := context.Background()

	t.Run("too few args", func(t *testing.T) {
		_, err := DecodeRows(ctx)
		if err == nil {
			t.Fatal("expected error for no args")
		}
	})

	t.Run("too many args", func(t *testing.T) {
		_, err := DecodeRows(ctx, runtime.NewString("a,b"), runtime.NewObject(), runtime.NewString("extra"))
		if err == nil {
			t.Fatal("expected error for too many args")
		}
	})

	t.Run("wrong arg type", func(t *testing.T) {
		_, err := DecodeRows(ctx, runtime.NewInt(42))
		if err == nil {
			t.Fatal("expected error for wrong arg type")
		}
	})

	t.Run("basic call", func(t *testing.T) {
		result, err := DecodeRows(ctx, runtime.NewString("a,b\nc,d"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		arr := mustRuntimeArray(t, result)
		assertRuntimeArrayLen(t, ctx, arr, 2)
	})

	t.Run("valid call with options", func(t *testing.T) {
		opts := runtime.NewObjectWith(map[string]runtime.Value{
			"delimiter":  runtime.NewString(";"),
			"inferTypes": runtime.True,
		})

		result, err := DecodeRows(ctx, runtime.NewString("1;2\n3;4"), opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		arr := mustRuntimeArray(t, result)
		assertRuntimeArrayLen(t, ctx, arr, 2)

		firstRow := mustArrayAtIndex(t, ctx, arr, 0)
		val, err := firstRow.At(ctx, runtime.Int(0))
		if err != nil {
			t.Fatalf("unexpected error getting first value: %v", err)
		}
		assertRuntimeIntValue(t, val, 1)
	})

	t.Run("wrong options type", func(t *testing.T) {
		_, err := DecodeRows(ctx, runtime.NewString("a,b"), runtime.NewString("not-a-map"))
		if err == nil {
			t.Fatal("expected error for wrong options type")
		}
	})

	t.Run("invalid delimiter option returns error", func(t *testing.T) {
		opts := runtime.NewObjectWith(map[string]runtime.Value{
			"delimiter": runtime.NewString("||"),
		})

		_, err := DecodeRows(ctx, runtime.NewString("a,b\nc,d"), opts)
		if err == nil {
			t.Fatal("expected error for invalid delimiter")
		}
	})
}
