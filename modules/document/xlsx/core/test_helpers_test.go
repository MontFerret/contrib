package core

import (
	"context"
	"testing"

	ferretfs "github.com/MontFerret/ferret/v2/pkg/fs"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type closeable interface {
	Close() error
}

func assertValue(t *testing.T, actual, expected runtime.Value) {
	t.Helper()

	if runtime.CompareValues(actual, expected) != 0 {
		t.Fatalf("expected %T(%q), got %T(%q)", expected, expected.String(), actual, actual.String())
	}
}

func listRows(t *testing.T, ctx context.Context, list runtime.List) [][]runtime.Value {
	t.Helper()

	length, err := list.Length(ctx)
	if err != nil {
		t.Fatalf("unexpected length error: %v", err)
	}

	rows := make([][]runtime.Value, 0, length)
	for idx := runtime.ZeroInt; idx < length; idx++ {
		rowValue, err := list.At(ctx, idx)
		if err != nil {
			t.Fatalf("unexpected row error: %v", err)
		}

		row, ok := rowValue.(runtime.List)
		if !ok {
			t.Fatalf("expected row list, got %T", rowValue)
		}

		rowLength, err := row.Length(ctx)
		if err != nil {
			t.Fatalf("unexpected row length error: %v", err)
		}

		values := make([]runtime.Value, 0, rowLength)
		for col := runtime.ZeroInt; col < rowLength; col++ {
			value, err := row.At(ctx, col)
			if err != nil {
				t.Fatalf("unexpected cell error: %v", err)
			}

			values = append(values, value)
		}

		rows = append(rows, values)
	}

	return rows
}

func objectField(t *testing.T, ctx context.Context, value runtime.Value, key string) runtime.Value {
	t.Helper()

	obj, ok := value.(runtime.Map)
	if !ok {
		t.Fatalf("expected object, got %T", value)
	}

	out, err := obj.Get(ctx, runtime.NewString(key))
	if err != nil {
		t.Fatalf("unexpected field error: %v", err)
	}

	return out
}

func testFSContext(t *testing.T, readOnly bool) (context.Context, string) {
	t.Helper()

	root := t.TempDir()
	filesystem, err := ferretfs.New(ferretfs.WithRoot(root), ferretfs.WithReadOnly(readOnly))
	if err != nil {
		t.Fatalf("unexpected filesystem error: %v", err)
	}
	if filesystemCloser, ok := filesystem.(closeable); ok {
		t.Cleanup(func() {
			if err := filesystemCloser.Close(); err != nil {
				t.Fatalf("unexpected filesystem close error: %v", err)
			}
		})
	}

	return ferretfs.WithFileSystem(context.Background(), filesystem), root
}
