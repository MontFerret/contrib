package lib

import (
	"context"
	"testing"

	"github.com/MontFerret/contrib/modules/csv/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestDecodeRowsStreamLib(t *testing.T) {
	ctx := context.Background()

	t.Run("too few args", func(t *testing.T) {
		_, err := DecodeRowsStream(ctx)
		if err == nil {
			t.Fatal("expected error for no args")
		}
	})

	t.Run("too many args", func(t *testing.T) {
		_, err := DecodeRowsStream(ctx, runtime.NewString("a,b"), runtime.NewObject(), runtime.NewString("extra"))
		if err == nil {
			t.Fatal("expected error for too many args")
		}
	})

	t.Run("iterates row arrays with options", func(t *testing.T) {
		opts := runtime.NewObjectWith(map[string]runtime.Value{
			"skipEmpty":  runtime.True,
			"inferTypes": runtime.True,
			"strict":     runtime.False,
		})

		result, err := DecodeRowsStream(ctx, runtime.NewString("42,3.14,true\n,\n7,2.71,false"), opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		iter := mustIterate(t, ctx, result)

		first, _, err := iter.Next(ctx)
		if err != nil {
			t.Fatalf("unexpected error reading first row: %v", err)
		}

		firstRow := mustRuntimeArray(t, first)
		assertRuntimeArrayLen(t, ctx, firstRow, 3)

		val, err := firstRow.At(ctx, runtime.Int(0))
		if err != nil {
			t.Fatalf("unexpected error getting first value: %v", err)
		}
		assertRuntimeIntValue(t, val, 42)

		val, err = firstRow.At(ctx, runtime.Int(2))
		if err != nil {
			t.Fatalf("unexpected error getting boolean value: %v", err)
		}
		if val != runtime.True {
			t.Fatalf("expected True, got %v", val)
		}

		second, _, err := iter.Next(ctx)
		if err != nil {
			t.Fatalf("unexpected error reading second row: %v", err)
		}

		secondRow := mustRuntimeArray(t, second)
		val, err = secondRow.At(ctx, runtime.Int(0))
		if err != nil {
			t.Fatalf("unexpected error getting second-row value: %v", err)
		}
		assertRuntimeIntValue(t, val, 7)
	})

	t.Run("accepts binary input", func(t *testing.T) {
		result, err := DecodeRowsStream(ctx, runtime.NewBinary([]byte("a,b\nc,d")))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		iter := mustIterate(t, ctx, result)
		first, _, err := iter.Next(ctx)
		if err != nil {
			t.Fatalf("unexpected error reading first row: %v", err)
		}

		firstRow := mustRuntimeArray(t, first)
		assertRuntimeArrayLen(t, ctx, firstRow, 2)
	})

	t.Run("propagates strict row errors during iteration", func(t *testing.T) {
		result, err := DecodeRowsStream(ctx, runtime.NewString("a,b\nc,d,e"))
		if err != nil {
			t.Fatalf("unexpected constructor error: %v", err)
		}

		iter := mustIterate(t, ctx, result)

		_, _, err = iter.Next(ctx)
		if err != nil {
			t.Fatalf("unexpected error on first row: %v", err)
		}

		_, _, err = iter.Next(ctx)
		if err == nil {
			t.Fatal("expected strict decode error on second row")
		}

		if _, ok := err.(*core.Error); !ok {
			t.Fatalf("expected *core.CSVError, got %T", err)
		}
	})

	t.Run("invalid comment option returns error", func(t *testing.T) {
		opts := runtime.NewObjectWith(map[string]runtime.Value{
			"comment": runtime.NewString("##"),
		})

		_, err := DecodeRowsStream(ctx, runtime.NewString("a,b\nc,d"), opts)
		if err == nil {
			t.Fatal("expected error for invalid comment")
		}
	})

	t.Run("rejects non-text input", func(t *testing.T) {
		_, err := DecodeRowsStream(ctx, runtime.NewObject())
		if err == nil {
			t.Fatal("expected error for non-text input")
		}
	})

	t.Run("wrong options type", func(t *testing.T) {
		_, err := DecodeRowsStream(ctx, runtime.NewString("a,b\nc,d"), runtime.NewString("not-a-map"))
		if err == nil {
			t.Fatal("expected error for wrong options type")
		}
	})
}
