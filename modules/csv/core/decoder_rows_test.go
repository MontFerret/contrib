package core

import (
	"context"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestDecodeRows(t *testing.T) {
	ctx := context.Background()

	t.Run("basic CSV includes all rows", func(t *testing.T) {
		opts := DefaultOptions()
		result, err := DecodeRows(ctx, runtime.NewString("name,age\nAlice,30\nBob,25"), opts)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		arr := mustArray(t, result)
		assertArrayLen(t, ctx, arr, 3) // header + 2 data rows

		row0 := mustArrayAt(t, ctx, arr, 0)
		assertArrayLen(t, ctx, row0, 2)
		assertArrayValue(t, ctx, row0, 0, "name")
		assertArrayValue(t, ctx, row0, 1, "age")

		row1 := mustArrayAt(t, ctx, arr, 1)
		assertArrayValue(t, ctx, row1, 0, "Alice")
		assertArrayValue(t, ctx, row1, 1, "30")
	})

	t.Run("skip empty rows", func(t *testing.T) {
		opts := DefaultOptions()
		opts.SkipEmpty = true

		result, err := DecodeRows(ctx, runtime.NewString("a,b\n,\nc,d"), opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		arr := mustArray(t, result)
		assertArrayLen(t, ctx, arr, 2) // "a,b" and "c,d", empty row skipped
	})

	t.Run("strict mode inconsistent columns error", func(t *testing.T) {
		opts := DefaultOptions()
		opts.Strict = true

		_, err := DecodeRows(ctx, runtime.NewString("a,b\nc,d,e"), opts)
		if err == nil {
			t.Fatal("expected error for inconsistent columns")
		}
	})

	t.Run("type inference in row mode", func(t *testing.T) {
		opts := DefaultOptions()
		opts.InferTypes = true

		result, err := DecodeRows(ctx, runtime.NewString("42,3.14,true,hello"), opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		arr := mustArray(t, result)
		row0 := mustArrayAt(t, ctx, arr, 0)

		val0, _ := row0.At(ctx, 0)
		assertRuntimeInt(t, val0, 42)

		val1, _ := row0.At(ctx, 1)
		assertRuntimeFloat(t, val1, 3.14)

		val2, _ := row0.At(ctx, 2)
		if val2 != runtime.True {
			t.Fatalf("expected True, got %v", val2)
		}

		val3, _ := row0.At(ctx, 3)
		if val3.String() != "hello" {
			t.Fatalf("expected 'hello', got %v", val3)
		}
	})

	t.Run("empty input", func(t *testing.T) {
		opts := DefaultOptions()
		result, err := DecodeRows(ctx, runtime.NewString(""), opts)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		arr := mustArray(t, result)
		assertArrayLen(t, ctx, arr, 0)
	})

	t.Run("relaxed mode preserves extra fields", func(t *testing.T) {
		opts := DefaultOptions()
		opts.Strict = false

		result, err := DecodeRows(ctx, runtime.NewString("a,b\nc,d,e"), opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		arr := mustArray(t, result)
		row1 := mustArrayAt(t, ctx, arr, 1)
		assertArrayLen(t, ctx, row1, 3) // extra field preserved
		assertArrayValue(t, ctx, row1, 2, "e")
	})

	t.Run("invalid multi-rune delimiter returns error", func(t *testing.T) {
		opts := DefaultOptions()
		opts.Delimiter = "||"

		_, err := DecodeRows(ctx, runtime.NewString("a,b\nc,d"), opts)
		if err == nil {
			t.Fatal("expected error for invalid delimiter")
		}
	})

	t.Run("invalid multi-rune comment returns error", func(t *testing.T) {
		opts := DefaultOptions()
		opts.Comment = "##"

		_, err := DecodeRows(ctx, runtime.NewString("a,b\n# comment\nc,d"), opts)
		if err == nil {
			t.Fatal("expected error for invalid comment")
		}
	})

	t.Run("invalid comment rune returns error", func(t *testing.T) {
		opts := DefaultOptions()
		opts.Comment = "\""

		_, err := DecodeRows(ctx, runtime.NewString("a,b\nc,d"), opts)
		if err == nil {
			t.Fatal("expected error for invalid comment")
		}
	})
}

func mustArrayAt(t *testing.T, ctx context.Context, arr *runtime.Array, idx int) *runtime.Array {
	t.Helper()

	val, err := arr.At(ctx, runtime.Int(idx))
	if err != nil {
		t.Fatalf("unexpected error at index %d: %v", idx, err)
	}

	inner, ok := val.(*runtime.Array)
	if !ok {
		t.Fatalf("expected *runtime.Array at index %d, got %T", idx, val)
	}

	return inner
}

func assertArrayValue(t *testing.T, ctx context.Context, arr *runtime.Array, idx int, expected string) {
	t.Helper()

	val, err := arr.At(ctx, runtime.Int(idx))
	if err != nil {
		t.Fatalf("unexpected error at index %d: %v", idx, err)
	}

	if val.String() != expected {
		t.Fatalf("index %d: expected %q, got %q", idx, expected, val.String())
	}
}
