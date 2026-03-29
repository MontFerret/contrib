package core

import (
	"errors"
	"testing"
)

func TestResolveHeaders(t *testing.T) {
	t.Run("header true", func(t *testing.T) {
		opts := Options{Header: true, Strict: true}
		headers, consumed, err := ResolveHeaders([]string{"name", "age", "city"}, opts)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !consumed {
			t.Fatal("expected consumed=true")
		}

		assertStringSlice(t, headers, []string{"name", "age", "city"})
	})

	t.Run("header false with columns", func(t *testing.T) {
		opts := Options{Header: false, Columns: []string{"first", "last"}}
		headers, consumed, err := ResolveHeaders([]string{"Alice", "Smith"}, opts)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if consumed {
			t.Fatal("expected consumed=false")
		}

		assertStringSlice(t, headers, []string{"first", "last"})
	})

	t.Run("header false no columns auto-generate", func(t *testing.T) {
		opts := Options{Header: false}
		headers, consumed, err := ResolveHeaders([]string{"a", "b", "c"}, opts)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if consumed {
			t.Fatal("expected consumed=false")
		}

		assertStringSlice(t, headers, []string{"col1", "col2", "col3"})
	})

	t.Run("header true with columns is error", func(t *testing.T) {
		opts := Options{Header: true, Columns: []string{"a", "b"}}
		_, _, err := ResolveHeaders([]string{"x", "y"}, opts)

		if !errors.Is(err, ErrHeaderColumnConflict) {
			t.Fatalf("expected ErrHeaderColumnConflict, got %v", err)
		}
	})

	t.Run("duplicate headers strict error", func(t *testing.T) {
		opts := Options{Header: true, Strict: true}
		_, _, err := ResolveHeaders([]string{"name", "age", "name"}, opts)

		if err == nil {
			t.Fatal("expected error for duplicate headers")
		}
	})

	t.Run("duplicate headers relaxed auto-rename", func(t *testing.T) {
		opts := Options{Header: true, Strict: false}
		headers, _, err := ResolveHeaders([]string{"name", "age", "name"}, opts)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		assertStringSlice(t, headers, []string{"name", "age", "name_2"})
	})

	t.Run("empty header strict error", func(t *testing.T) {
		opts := Options{Header: true, Strict: true}
		_, _, err := ResolveHeaders([]string{"name", "", "age"}, opts)

		if err == nil {
			t.Fatal("expected error for empty header")
		}
	})

	t.Run("empty header relaxed synthesize", func(t *testing.T) {
		opts := Options{Header: true, Strict: false}
		headers, _, err := ResolveHeaders([]string{"name", "", "age"}, opts)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		assertStringSlice(t, headers, []string{"name", "col2", "age"})
	})

	t.Run("synthesized headers stay unique in relaxed mode", func(t *testing.T) {
		opts := Options{Header: true, Strict: false}
		headers, _, err := ResolveHeaders([]string{"col2", "", "col2"}, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		assertStringSlice(t, headers, []string{"col2", "col2_2", "col2_3"})
	})

	t.Run("renamed headers avoid existing literal suffix collisions", func(t *testing.T) {
		opts := Options{Header: true, Strict: false}
		headers, _, err := ResolveHeaders([]string{"a", "a_2", "a"}, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		assertStringSlice(t, headers, []string{"a", "a_2", "a_3"})
	})

	t.Run("renamed headers skip multiple occupied suffixes", func(t *testing.T) {
		opts := Options{Header: true, Strict: false}
		headers, _, err := ResolveHeaders([]string{"a", "a_2", "a_3", "a"}, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		assertStringSlice(t, headers, []string{"a", "a_2", "a_3", "a_4"})
	})
}

func assertStringSlice(t *testing.T, got, want []string) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("length mismatch: got %d, want %d", len(got), len(want))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("index %d: got %q, want %q", i, got[i], want[i])
		}
	}
}
