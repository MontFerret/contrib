package lib

import (
	"context"
	"io"
	"testing"

	"github.com/MontFerret/contrib/modules/csv/types"
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

	t.Run("camelCase options are decoded", func(t *testing.T) {
		opts := runtime.NewObjectWith(map[string]runtime.Value{
			"skipEmpty":  runtime.True,
			"inferTypes": runtime.True,
			"nullValues": runtime.NewArrayWith(runtime.NewString("null")),
		})

		result, err := Decode(ctx, runtime.NewString("name,age\nAlice,30\n,\nBob,null"), opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		arr := mustRuntimeArray(t, result)
		assertRuntimeArrayLen(t, ctx, arr, 2)

		first := mustObjectAtIndex(t, ctx, arr, 0)
		age, err := first.Get(ctx, runtime.NewString("age"))
		if err != nil {
			t.Fatalf("unexpected error getting age: %v", err)
		}
		assertRuntimeIntValue(t, age, 30)

		second := mustObjectAtIndex(t, ctx, arr, 1)
		age, err = second.Get(ctx, runtime.NewString("age"))
		if err != nil {
			t.Fatalf("unexpected error getting null age: %v", err)
		}
		if age != runtime.None {
			t.Fatalf("expected null age to decode as None, got %v", age)
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

func TestDecodeStreamLib(t *testing.T) {
	ctx := context.Background()

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

		if _, ok := err.(*types.CSVError); !ok {
			t.Fatalf("expected *types.CSVError, got %T", err)
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
}

func TestDecodeRowsStreamLib(t *testing.T) {
	ctx := context.Background()

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

		if _, ok := err.(*types.CSVError); !ok {
			t.Fatalf("expected *types.CSVError, got %T", err)
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
}

func mustIterate(t *testing.T, ctx context.Context, val runtime.Value) runtime.Iterator {
	t.Helper()

	iterable, ok := val.(runtime.Iterable)
	if !ok {
		t.Fatalf("expected runtime.Iterable, got %T", val)
	}

	iter, err := iterable.Iterate(ctx)
	if err != nil {
		t.Fatalf("unexpected iterate error: %v", err)
	}

	return iter
}

func mustRuntimeArray(t *testing.T, val runtime.Value) *runtime.Array {
	t.Helper()

	arr, ok := val.(*runtime.Array)
	if !ok {
		t.Fatalf("expected *runtime.Array, got %T", val)
	}

	return arr
}

func mustRuntimeObject(t *testing.T, val runtime.Value) *runtime.Object {
	t.Helper()

	obj, ok := val.(*runtime.Object)
	if !ok {
		t.Fatalf("expected *runtime.Object, got %T", val)
	}

	return obj
}

func mustObjectAtIndex(t *testing.T, ctx context.Context, arr *runtime.Array, idx int) *runtime.Object {
	t.Helper()

	val, err := arr.At(ctx, runtime.Int(idx))
	if err != nil {
		t.Fatalf("unexpected error at index %d: %v", idx, err)
	}

	return mustRuntimeObject(t, val)
}

func assertRuntimeArrayLen(t *testing.T, ctx context.Context, arr *runtime.Array, expected int) {
	t.Helper()

	length, err := arr.Length(ctx)
	if err != nil {
		t.Fatalf("unexpected length error: %v", err)
	}

	if int(length) != expected {
		t.Fatalf("expected length %d, got %d", expected, int(length))
	}
}

func assertRuntimeIntValue(t *testing.T, val runtime.Value, expected int64) {
	t.Helper()

	i, ok := val.(runtime.Int)
	if !ok {
		t.Fatalf("expected runtime.Int, got %T", val)
	}

	if int64(i) != expected {
		t.Fatalf("expected %d, got %d", expected, int64(i))
	}
}

func assertObjectField(t *testing.T, ctx context.Context, obj *runtime.Object, key, expected string) {
	t.Helper()

	val, err := obj.Get(ctx, runtime.NewString(key))
	if err != nil {
		t.Fatalf("unexpected error getting %q: %v", key, err)
	}

	if val.String() != expected {
		t.Fatalf("field %q: expected %q, got %q", key, expected, val.String())
	}
}
