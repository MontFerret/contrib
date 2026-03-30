package core

import (
	"context"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestDecode(t *testing.T) {
	ctx := context.Background()

	t.Run("basic CSV with header", func(t *testing.T) {
		opts := DefaultOptions()
		result, err := Decode(ctx, runtime.NewString("name,age\nAlice,30\nBob,25"), opts)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		arr := mustArray(t, result)
		assertArrayLen(t, ctx, arr, 2)

		obj0 := mustObjectAt(t, ctx, arr, 0)
		assertField(t, ctx, obj0, "name", "Alice")
		assertField(t, ctx, obj0, "age", "30")

		obj1 := mustObjectAt(t, ctx, arr, 1)
		assertField(t, ctx, obj1, "name", "Bob")
		assertField(t, ctx, obj1, "age", "25")
	})

	t.Run("no header explicit columns", func(t *testing.T) {
		opts := DefaultOptions()
		opts.Header = false
		opts.Columns = []string{"first", "last"}

		result, err := Decode(ctx, runtime.NewString("Alice,Smith\nBob,Jones"), opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		arr := mustArray(t, result)
		assertArrayLen(t, ctx, arr, 2)

		obj0 := mustObjectAt(t, ctx, arr, 0)
		assertField(t, ctx, obj0, "first", "Alice")
		assertField(t, ctx, obj0, "last", "Smith")
	})

	t.Run("no header no columns auto-generate", func(t *testing.T) {
		opts := DefaultOptions()
		opts.Header = false
		opts.Strict = false

		result, err := Decode(ctx, runtime.NewString("Alice,30\nBob,25"), opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		arr := mustArray(t, result)
		obj0 := mustObjectAt(t, ctx, arr, 0)
		assertField(t, ctx, obj0, "col1", "Alice")
		assertField(t, ctx, obj0, "col2", "30")
	})

	t.Run("custom delimiter semicolon", func(t *testing.T) {
		opts := DefaultOptions()
		opts.Delimiter = ";"

		result, err := Decode(ctx, runtime.NewString("name;age\nAlice;30"), opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		arr := mustArray(t, result)
		obj0 := mustObjectAt(t, ctx, arr, 0)
		assertField(t, ctx, obj0, "name", "Alice")
		assertField(t, ctx, obj0, "age", "30")
	})

	t.Run("custom delimiter tab", func(t *testing.T) {
		opts := DefaultOptions()
		opts.Delimiter = "\t"

		result, err := Decode(ctx, runtime.NewString("name\tage\nAlice\t30"), opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		arr := mustArray(t, result)
		obj0 := mustObjectAt(t, ctx, arr, 0)
		assertField(t, ctx, obj0, "name", "Alice")
		assertField(t, ctx, obj0, "age", "30")
	})

	t.Run("invalid delimiter rune returns error before decoding", func(t *testing.T) {
		opts := DefaultOptions()
		opts.Delimiter = "\n"

		_, err := Decode(ctx, runtime.NewString("name,age\nAlice,30"), opts)
		if err == nil {
			t.Fatal("expected error for invalid delimiter")
		}
	})

	t.Run("comment equal to delimiter returns error before decoding", func(t *testing.T) {
		opts := DefaultOptions()
		opts.Comment = ","

		_, err := Decode(ctx, runtime.NewString("name,age\nAlice,30"), opts)
		if err == nil {
			t.Fatal("expected error for comment equal to delimiter")
		}
	})

	t.Run("skip empty rows", func(t *testing.T) {
		opts := DefaultOptions()
		opts.SkipEmpty = true

		result, err := Decode(ctx, runtime.NewString("name,age\nAlice,30\n,\nBob,25"), opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		arr := mustArray(t, result)
		assertArrayLen(t, ctx, arr, 2)
	})

	t.Run("skip leading empty record before header", func(t *testing.T) {
		opts := DefaultOptions()
		opts.SkipEmpty = true

		result, err := Decode(ctx, runtime.NewString(",\nname,age\nAlice,30"), opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		arr := mustArray(t, result)
		assertArrayLen(t, ctx, arr, 1)
		assertField(t, ctx, mustObjectAt(t, ctx, arr, 0), "name", "Alice")
		assertField(t, ctx, mustObjectAt(t, ctx, arr, 0), "age", "30")
	})

	t.Run("skip leading empty record before auto-generated columns", func(t *testing.T) {
		opts := DefaultOptions()
		opts.Header = false
		opts.SkipEmpty = true
		opts.Strict = true

		result, err := Decode(ctx, runtime.NewString(",\nAlice,30\nBob,25"), opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		arr := mustArray(t, result)
		assertArrayLen(t, ctx, arr, 2)
		assertField(t, ctx, mustObjectAt(t, ctx, arr, 0), "col1", "Alice")
		assertField(t, ctx, mustObjectAt(t, ctx, arr, 0), "col2", "30")
	})

	t.Run("strict mode first headerless row uses original row number", func(t *testing.T) {
		opts := DefaultOptions()
		opts.Header = false
		opts.Strict = true
		opts.Columns = []string{"first", "last"}

		_, err := Decode(ctx, runtime.NewString("Alice,Smith,extra\nBob,Jones"), opts)
		if err == nil {
			t.Fatal("expected error for inconsistent columns")
		}

		csvErr, ok := err.(*Error)
		if !ok {
			t.Fatalf("expected CSVError, got %T", err)
		}

		if csvErr.Row != 1 {
			t.Fatalf("expected row 1, got %d", csvErr.Row)
		}
	})

	t.Run("strict mode headerless row preserves skipped leading empty record numbering", func(t *testing.T) {
		opts := DefaultOptions()
		opts.Header = false
		opts.SkipEmpty = true
		opts.Strict = true
		opts.Columns = []string{"first", "last"}

		_, err := Decode(ctx, runtime.NewString(",\nAlice,Smith,extra\nBob,Jones"), opts)
		if err == nil {
			t.Fatal("expected error for inconsistent columns")
		}

		csvErr, ok := err.(*Error)
		if !ok {
			t.Fatalf("expected CSVError, got %T", err)
		}

		if csvErr.Row != 2 {
			t.Fatalf("expected row 2, got %d", csvErr.Row)
		}
	})

	t.Run("strict mode inconsistent columns error", func(t *testing.T) {
		opts := DefaultOptions()
		opts.Strict = true

		_, err := Decode(ctx, runtime.NewString("name,age\nAlice,30\nBob,25,extra"), opts)
		if err == nil {
			t.Fatal("expected error for inconsistent columns")
		}

		csvErr, ok := err.(*Error)
		if !ok {
			t.Fatalf("expected CSVError, got %T", err)
		}

		if csvErr.Row != 3 {
			t.Fatalf("expected row 3, got %d", csvErr.Row)
		}
	})

	t.Run("strict mode error row includes skipped leading empty record", func(t *testing.T) {
		opts := DefaultOptions()
		opts.SkipEmpty = true
		opts.Strict = true

		_, err := Decode(ctx, runtime.NewString(",\nname,age\nAlice,30,extra"), opts)
		if err == nil {
			t.Fatal("expected error for inconsistent columns")
		}

		csvErr, ok := err.(*Error)
		if !ok {
			t.Fatalf("expected CSVError, got %T", err)
		}

		if csvErr.Row != 3 {
			t.Fatalf("expected row 3, got %d", csvErr.Row)
		}
	})

	t.Run("relaxed mode short row fills null", func(t *testing.T) {
		opts := DefaultOptions()
		opts.Strict = false

		result, err := Decode(ctx, runtime.NewString("name,age,city\nAlice,30"), opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		arr := mustArray(t, result)
		obj0 := mustObjectAt(t, ctx, arr, 0)

		val, err := obj0.Get(ctx, runtime.NewString("city"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if val != runtime.None {
			t.Fatalf("expected None for missing field, got %v", val)
		}
	})

	t.Run("relaxed mode extra fields ignored", func(t *testing.T) {
		opts := DefaultOptions()
		opts.Strict = false

		result, err := Decode(ctx, runtime.NewString("name,age\nAlice,30,extra"), opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		arr := mustArray(t, result)
		obj0 := mustObjectAt(t, ctx, arr, 0)

		// Should only have name and age
		assertField(t, ctx, obj0, "name", "Alice")
		assertField(t, ctx, obj0, "age", "30")
	})

	t.Run("comment lines skipped", func(t *testing.T) {
		opts := DefaultOptions()
		opts.Comment = "#"

		result, err := Decode(ctx, runtime.NewString("name,age\n# this is a comment\nAlice,30"), opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		arr := mustArray(t, result)
		assertArrayLen(t, ctx, arr, 1)
		assertField(t, ctx, mustObjectAt(t, ctx, arr, 0), "name", "Alice")
	})

	t.Run("type inference", func(t *testing.T) {
		opts := DefaultOptions()
		opts.InferTypes = true

		result, err := Decode(ctx, runtime.NewString("name,age,active\nAlice,30,true"), opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		arr := mustArray(t, result)
		obj0 := mustObjectAt(t, ctx, arr, 0)

		age, err := obj0.Get(ctx, runtime.NewString("age"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		assertRuntimeInt(t, age, 30)

		active, err := obj0.Get(ctx, runtime.NewString("active"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if active != runtime.True {
			t.Fatalf("expected True, got %v", active)
		}
	})

	t.Run("empty input", func(t *testing.T) {
		opts := DefaultOptions()
		result, err := Decode(ctx, runtime.NewString(""), opts)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		arr := mustArray(t, result)
		assertArrayLen(t, ctx, arr, 0)
	})

	t.Run("all leading empty records with skip empty returns empty result", func(t *testing.T) {
		opts := DefaultOptions()
		opts.SkipEmpty = true

		result, err := Decode(ctx, runtime.NewString(",\n,\n"), opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		arr := mustArray(t, result)
		assertArrayLen(t, ctx, arr, 0)
	})

	t.Run("quoted fields", func(t *testing.T) {
		opts := DefaultOptions()
		result, err := Decode(ctx, runtime.NewString("name,quote\nAlice,\"hello, world\"\nBob,\"He said \"\"hi\"\"\""), opts)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		arr := mustArray(t, result)
		assertArrayLen(t, ctx, arr, 2)
		assertField(t, ctx, mustObjectAt(t, ctx, arr, 0), "quote", "hello, world")
		assertField(t, ctx, mustObjectAt(t, ctx, arr, 1), "quote", "He said \"hi\"")
	})

	t.Run("header true with columns is error", func(t *testing.T) {
		opts := DefaultOptions()
		opts.Columns = []string{"a", "b"}

		_, err := Decode(ctx, runtime.NewString("x,y\n1,2"), opts)
		if err == nil {
			t.Fatal("expected error for header+columns conflict")
		}
	})

	t.Run("invalid multi-rune delimiter returns error", func(t *testing.T) {
		opts := DefaultOptions()
		opts.Delimiter = "||"

		_, err := Decode(ctx, runtime.NewString("name,age\nAlice,30"), opts)
		if err == nil {
			t.Fatal("expected error for invalid delimiter")
		}
	})

	t.Run("invalid multi-rune comment returns error", func(t *testing.T) {
		opts := DefaultOptions()
		opts.Comment = "##"

		_, err := Decode(ctx, runtime.NewString("name,age\n# comment\nAlice,30"), opts)
		if err == nil {
			t.Fatal("expected error for invalid comment")
		}
	})
}

// Test helpers

func mustArray(t *testing.T, val runtime.Value) *runtime.Array {
	t.Helper()

	arr, ok := val.(*runtime.Array)
	if !ok {
		t.Fatalf("expected *runtime.Array, got %T", val)
	}

	return arr
}

func assertArrayLen(t *testing.T, ctx context.Context, arr *runtime.Array, expected int) {
	t.Helper()

	length, err := arr.Length(ctx)
	if err != nil {
		t.Fatalf("unexpected error getting length: %v", err)
	}

	if int(length) != expected {
		t.Fatalf("expected length %d, got %d", expected, int(length))
	}
}

func mustObjectAt(t *testing.T, ctx context.Context, arr *runtime.Array, idx int) *runtime.Object {
	t.Helper()

	val, err := arr.At(ctx, runtime.Int(idx))
	if err != nil {
		t.Fatalf("unexpected error at index %d: %v", idx, err)
	}

	obj, ok := val.(*runtime.Object)
	if !ok {
		t.Fatalf("expected *runtime.Object at index %d, got %T", idx, val)
	}

	return obj
}

func assertField(t *testing.T, ctx context.Context, obj *runtime.Object, key, expected string) {
	t.Helper()

	val, err := obj.Get(ctx, runtime.NewString(key))
	if err != nil {
		t.Fatalf("unexpected error getting %q: %v", key, err)
	}

	if val.String() != expected {
		t.Fatalf("field %q: expected %q, got %q", key, expected, val.String())
	}
}
