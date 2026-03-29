package core

import (
	"context"
	"encoding/csv"
	"errors"
	"strings"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestEncode(t *testing.T) {
	ctx := context.Background()

	t.Run("array of objects with header", func(t *testing.T) {
		opts := DefaultOptions()
		data := runtime.NewArrayWith(
			runtime.NewObjectWith(map[string]runtime.Value{
				"name": runtime.NewString("Alice"),
				"age":  runtime.NewInt(30),
			}),
			runtime.NewObjectWith(map[string]runtime.Value{
				"name": runtime.NewString("Bob"),
				"age":  runtime.NewInt(25),
			}),
		)

		result, err := Encode(ctx, data, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		lines := strings.Split(strings.TrimSpace(result.Text), "\n")

		if result.Rows != 2 {
			t.Fatalf("expected 2 rows, got %d", result.Rows)
		}

		// Header should be present
		if len(lines) != 3 {
			t.Fatalf("expected 3 lines (header + 2 data), got %d", len(lines))
		}

		// Verify header contains both fields
		header := lines[0]
		if !strings.Contains(header, "name") || !strings.Contains(header, "age") {
			t.Fatalf("header should contain 'name' and 'age', got %q", header)
		}
	})

	t.Run("array of objects without header", func(t *testing.T) {
		opts := DefaultOptions()
		opts.Header = false

		data := runtime.NewArrayWith(
			runtime.NewObjectWith(map[string]runtime.Value{
				"name": runtime.NewString("Alice"),
			}),
		)

		result, err := Encode(ctx, data, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		lines := strings.Split(strings.TrimSpace(result.Text), "\n")

		if len(lines) != 1 {
			t.Fatalf("expected 1 line (no header), got %d", len(lines))
		}
	})

	t.Run("array of objects with columns option", func(t *testing.T) {
		opts := DefaultOptions()
		opts.Columns = []string{"age", "name"} // explicit order

		data := runtime.NewArrayWith(
			runtime.NewObjectWith(map[string]runtime.Value{
				"name": runtime.NewString("Alice"),
				"age":  runtime.NewInt(30),
			}),
		)

		result, err := Encode(ctx, data, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		lines := strings.Split(strings.TrimSpace(result.Text), "\n")

		if lines[0] != "age,name" {
			t.Fatalf("expected header 'age,name', got %q", lines[0])
		}

		if lines[1] != "30,Alice" {
			t.Fatalf("expected data '30,Alice', got %q", lines[1])
		}
	})

	t.Run("array of arrays", func(t *testing.T) {
		opts := DefaultOptions()
		data := runtime.NewArrayWith(
			runtime.NewArrayWith(runtime.NewString("name"), runtime.NewString("age")),
			runtime.NewArrayWith(runtime.NewString("Alice"), runtime.NewInt(30)),
		)

		result, err := Encode(ctx, data, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		lines := strings.Split(strings.TrimSpace(result.Text), "\n")

		if len(lines) != 2 {
			t.Fatalf("expected 2 lines, got %d", len(lines))
		}

		if lines[0] != "name,age" {
			t.Fatalf("expected 'name,age', got %q", lines[0])
		}

		if lines[1] != "Alice,30" {
			t.Fatalf("expected 'Alice,30', got %q", lines[1])
		}

		if result.Rows != 2 {
			t.Fatalf("expected 2 rows, got %d", result.Rows)
		}
	})

	t.Run("custom delimiter", func(t *testing.T) {
		opts := DefaultOptions()
		opts.Delimiter = ";"

		data := runtime.NewArrayWith(
			runtime.NewArrayWith(runtime.NewString("a"), runtime.NewString("b")),
		)

		result, err := Encode(ctx, data, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if strings.TrimSpace(result.Text) != "a;b" {
			t.Fatalf("expected 'a;b', got %q", strings.TrimSpace(result.Text))
		}
	})

	t.Run("invalid multi-rune delimiter returns error", func(t *testing.T) {
		opts := DefaultOptions()
		opts.Delimiter = "||"

		data := runtime.NewArrayWith(
			runtime.NewArrayWith(runtime.NewString("a"), runtime.NewString("b")),
		)

		_, err := Encode(ctx, data, opts)
		if err == nil {
			t.Fatal("expected error for invalid delimiter")
		}
	})

	t.Run("invalid delimiter rune returns error before encoding", func(t *testing.T) {
		opts := DefaultOptions()
		opts.Delimiter = "\""

		data := runtime.NewArrayWith(
			runtime.NewArrayWith(runtime.NewString("a"), runtime.NewString("b")),
		)

		_, err := Encode(ctx, data, opts)
		if err == nil {
			t.Fatal("expected error for invalid delimiter")
		}
	})

	t.Run("missing keys produce empty fields", func(t *testing.T) {
		opts := DefaultOptions()
		opts.Columns = []string{"name", "age", "city"}

		data := runtime.NewArrayWith(
			runtime.NewObjectWith(map[string]runtime.Value{
				"name": runtime.NewString("Alice"),
			}),
		)

		result, err := Encode(ctx, data, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		lines := strings.Split(strings.TrimSpace(result.Text), "\n")

		if lines[1] != "Alice,," {
			t.Fatalf("expected 'Alice,,', got %q", lines[1])
		}
	})

	t.Run("empty array", func(t *testing.T) {
		opts := DefaultOptions()
		data := runtime.NewArray(0)

		result, err := Encode(ctx, data, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Rows != 0 {
			t.Fatalf("expected 0 rows, got %d", result.Rows)
		}
	})

	t.Run("round trip decode then encode", func(t *testing.T) {
		input := "name,age\nAlice,30\nBob,25\n"
		opts := DefaultOptions()

		decoded, err := Decode(ctx, runtime.NewString(input), opts)
		if err != nil {
			t.Fatalf("decode error: %v", err)
		}

		encOpts := DefaultOptions()
		encOpts.Columns = []string{"name", "age"} // fix column order

		result, err := Encode(ctx, decoded, encOpts)
		if err != nil {
			t.Fatalf("encode error: %v", err)
		}

		if result.Text != input {
			t.Fatalf("round trip mismatch:\nexpected: %q\ngot:      %q", input, result.Text)
		}
	})

	t.Run("flush helper returns deferred writer error", func(t *testing.T) {
		writer := csv.NewWriter(failingWriter{})
		if err := writer.Write([]string{"a", "b"}); err != nil {
			t.Fatalf("unexpected write error before flush: %v", err)
		}

		err := flushWriter(writer)
		if err == nil {
			t.Fatal("expected flush error")
		}

		if !strings.Contains(err.Error(), "csv: failed to flush output") {
			t.Fatalf("expected flush error wrapper, got %v", err)
		}

		if !errors.Is(err, errFailingWriter) {
			t.Fatalf("expected wrapped failing writer error, got %v", err)
		}
	})
}

var errFailingWriter = errors.New("failing writer")

type failingWriter struct{}

func (failingWriter) Write(p []byte) (int, error) {
	return 0, errFailingWriter
}
