package lib

import (
	"context"
	"io"
	"testing"

	"github.com/MontFerret/contrib/modules/csv/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestDecodeStreamLib(t *testing.T) {
	ctx := context.Background()

	t.Run("too few args", func(t *testing.T) {
		_, err := DecodeStream(ctx)
		if err == nil {
			t.Fatal("expected error for no args")
		}
	})

	t.Run("too many args", func(t *testing.T) {
		_, err := DecodeStream(ctx, runtime.NewString("a,b"), runtime.NewObject(), runtime.NewString("extra"))
		if err == nil {
			t.Fatal("expected error for too many args")
		}
	})

	t.Run("iterates decoded objects with camelCase options", func(t *testing.T) {
		opts := runtime.NewObjectWith(map[string]runtime.Value{
			"skipEmpty":  runtime.True,
			"inferTypes": runtime.True,
		})

		result, err := DecodeStream(ctx, runtime.NewString("name,age\nAlice,30\n,\nBob,25"), opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		iter := mustIterate(t, ctx, result)

		first, _, err := iter.Next(ctx)
		if err != nil {
			t.Fatalf("unexpected error reading first row: %v", err)
		}

		firstObj := mustRuntimeObject(t, first)
		assertObjectField(t, ctx, firstObj, "name", "Alice")

		age, err := firstObj.Get(ctx, runtime.NewString("age"))
		if err != nil {
			t.Fatalf("unexpected error getting age: %v", err)
		}
		assertRuntimeIntValue(t, age, 30)

		second, _, err := iter.Next(ctx)
		if err != nil {
			t.Fatalf("unexpected error reading second row: %v", err)
		}

		secondObj := mustRuntimeObject(t, second)
		assertObjectField(t, ctx, secondObj, "name", "Bob")

		_, _, err = iter.Next(ctx)
		if err != io.EOF {
			t.Fatalf("expected EOF, got %v", err)
		}
	})

	t.Run("accepts binary input", func(t *testing.T) {
		result, err := DecodeStream(ctx, runtime.NewBinary([]byte("name,age\nAlice,30")))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		iter := mustIterate(t, ctx, result)
		first, _, err := iter.Next(ctx)
		if err != nil {
			t.Fatalf("unexpected error reading first row: %v", err)
		}

		firstObj := mustRuntimeObject(t, first)
		assertObjectField(t, ctx, firstObj, "name", "Alice")
	})

	t.Run("skips leading empty record before header", func(t *testing.T) {
		opts := runtime.NewObjectWith(map[string]runtime.Value{
			"skipEmpty": runtime.True,
		})

		result, err := DecodeStream(ctx, runtime.NewString(",\nname,age\nAlice,30"), opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		iter := mustIterate(t, ctx, result)
		first, _, err := iter.Next(ctx)
		if err != nil {
			t.Fatalf("unexpected error reading first row: %v", err)
		}

		firstObj := mustRuntimeObject(t, first)
		assertObjectField(t, ctx, firstObj, "name", "Alice")
		assertObjectField(t, ctx, firstObj, "age", "30")
	})

	t.Run("headerless first row keeps original iterator key", func(t *testing.T) {
		opts := runtime.NewObjectWith(map[string]runtime.Value{
			"header": runtime.False,
		})

		result, err := DecodeStream(ctx, runtime.NewString("Alice,30\nBob,25"), opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		iter := mustIterate(t, ctx, result)
		first, key, err := iter.Next(ctx)
		if err != nil {
			t.Fatalf("unexpected error reading first row: %v", err)
		}

		assertRuntimeIntValue(t, key, 1)

		firstObj := mustRuntimeObject(t, first)
		assertObjectField(t, ctx, firstObj, "col1", "Alice")
		assertObjectField(t, ctx, firstObj, "col2", "30")
	})

	t.Run("headerless first row preserves skipped leading empty record numbering", func(t *testing.T) {
		opts := runtime.NewObjectWith(map[string]runtime.Value{
			"header":    runtime.False,
			"skipEmpty": runtime.True,
		})

		result, err := DecodeStream(ctx, runtime.NewString(",\nAlice,30\nBob,25"), opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		iter := mustIterate(t, ctx, result)
		first, key, err := iter.Next(ctx)
		if err != nil {
			t.Fatalf("unexpected error reading first row: %v", err)
		}

		assertRuntimeIntValue(t, key, 2)

		firstObj := mustRuntimeObject(t, first)
		assertObjectField(t, ctx, firstObj, "col1", "Alice")
		assertObjectField(t, ctx, firstObj, "col2", "30")
	})

	t.Run("propagates strict decode errors during iteration", func(t *testing.T) {
		result, err := DecodeStream(ctx, runtime.NewString("name,age\nAlice,30\nBob,25,extra"))
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

		if _, ok := err.(*core.CSVError); !ok {
			t.Fatalf("expected *types.CSVError, got %T", err)
		}
	})

	t.Run("headerless strict error on first row keeps original row number", func(t *testing.T) {
		opts := runtime.NewObjectWith(map[string]runtime.Value{
			"header":  runtime.False,
			"strict":  runtime.True,
			"columns": runtime.NewArrayWith(runtime.NewString("first"), runtime.NewString("last")),
		})

		result, err := DecodeStream(ctx, runtime.NewString("Alice,Smith,extra\nBob,Jones"), opts)
		if err != nil {
			t.Fatalf("unexpected constructor error: %v", err)
		}

		iter := mustIterate(t, ctx, result)
		_, _, err = iter.Next(ctx)
		if err == nil {
			t.Fatal("expected strict decode error on first row")
		}

		csvErr, ok := err.(*core.CSVError)
		if !ok {
			t.Fatalf("expected *types.CSVError, got %T", err)
		}

		if csvErr.Row != 1 {
			t.Fatalf("expected row 1, got %d", csvErr.Row)
		}
	})

	t.Run("invalid delimiter option returns error", func(t *testing.T) {
		opts := runtime.NewObjectWith(map[string]runtime.Value{
			"delimiter": runtime.NewString("||"),
		})

		_, err := DecodeStream(ctx, runtime.NewString("name,age\nAlice,30"), opts)
		if err == nil {
			t.Fatal("expected error for invalid delimiter")
		}
	})

	t.Run("rejects non-text input", func(t *testing.T) {
		_, err := DecodeStream(ctx, runtime.NewInt(42))
		if err == nil {
			t.Fatal("expected error for non-text input")
		}
	})

	t.Run("wrong options type", func(t *testing.T) {
		_, err := DecodeStream(ctx, runtime.NewString("name,age\nAlice,30"), runtime.NewString("not-a-map"))
		if err == nil {
			t.Fatal("expected error for wrong options type")
		}
	})
}
