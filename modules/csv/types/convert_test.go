package types

import (
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestConvertValue(t *testing.T) {
	t.Run("string passthrough", func(t *testing.T) {
		opts := Options{}
		val := ConvertValue("hello", opts)

		if val.String() != "hello" {
			t.Fatalf("expected 'hello', got %q", val.String())
		}
	})

	t.Run("integer inference", func(t *testing.T) {
		opts := Options{InferTypes: true}
		val := ConvertValue("42", opts)
		assertRuntimeInt(t, val, 42)
	})

	t.Run("negative integer", func(t *testing.T) {
		opts := Options{InferTypes: true}
		val := ConvertValue("-10", opts)
		assertRuntimeInt(t, val, -10)
	})

	t.Run("float inference", func(t *testing.T) {
		opts := Options{InferTypes: true}
		val := ConvertValue("3.14", opts)
		assertRuntimeFloat(t, val, 3.14)
	})

	t.Run("boolean true", func(t *testing.T) {
		opts := Options{InferTypes: true}
		val := ConvertValue("true", opts)

		if val != runtime.True {
			t.Fatalf("expected True, got %v", val)
		}
	})

	t.Run("boolean false", func(t *testing.T) {
		opts := Options{InferTypes: true}
		val := ConvertValue("false", opts)

		if val != runtime.False {
			t.Fatalf("expected False, got %v", val)
		}
	})

	t.Run("boolean case insensitive", func(t *testing.T) {
		opts := Options{InferTypes: true}
		val := ConvertValue("TRUE", opts)

		if val != runtime.True {
			t.Fatalf("expected True, got %v", val)
		}
	})

	t.Run("no inference keeps strings", func(t *testing.T) {
		opts := Options{InferTypes: false}
		val := ConvertValue("42", opts)

		s, ok := val.(runtime.String)
		if !ok {
			t.Fatalf("expected runtime.String, got %T", val)
		}

		if s.String() != "42" {
			t.Fatalf("expected '42', got %q", s.String())
		}
	})

	t.Run("null values", func(t *testing.T) {
		opts := Options{NullValues: []string{"", "null", "N/A"}}

		for _, nv := range []string{"", "null", "N/A"} {
			val := ConvertValue(nv, opts)

			if val != runtime.None {
				t.Fatalf("expected None for %q, got %v", nv, val)
			}
		}
	})

	t.Run("null values no match", func(t *testing.T) {
		opts := Options{NullValues: []string{"null"}}
		val := ConvertValue("hello", opts)

		if val.String() != "hello" {
			t.Fatalf("expected 'hello', got %v", val)
		}
	})

	t.Run("trim", func(t *testing.T) {
		opts := Options{Trim: true}
		val := ConvertValue("  hello  ", opts)

		if val.String() != "hello" {
			t.Fatalf("expected 'hello', got %q", val.String())
		}
	})

	t.Run("trim with inference", func(t *testing.T) {
		opts := Options{Trim: true, InferTypes: true}
		val := ConvertValue("  42  ", opts)
		assertRuntimeInt(t, val, 42)
	})
}

func assertRuntimeInt(t *testing.T, val runtime.Value, expected int) {
	t.Helper()

	i, ok := val.(runtime.Int)
	if !ok {
		t.Fatalf("expected runtime.Int, got %T (%v)", val, val)
	}

	if int(i) != expected {
		t.Fatalf("expected %d, got %d", expected, int(i))
	}
}

func assertRuntimeFloat(t *testing.T, val runtime.Value, expected float64) {
	t.Helper()

	f, ok := val.(runtime.Float)
	if !ok {
		t.Fatalf("expected runtime.Float, got %T (%v)", val, val)
	}

	if float64(f) != expected {
		t.Fatalf("expected %f, got %f", expected, float64(f))
	}
}
