package core

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenMemoryDB(t *testing.T) {
	t.Parallel()

	db, err := Open(context.Background(), OpenOptions{Memory: boolPtr(true)})
	if err != nil {
		t.Fatalf("unexpected open error: %v", err)
	}
	defer db.Close()
}

func TestOpenFileBackedDB(t *testing.T) {
	t.Parallel()

	path := tempDBPath(t)
	db, err := Open(context.Background(), OpenOptions{Path: stringPtr(path)})
	if err != nil {
		t.Fatalf("unexpected open error: %v", err)
	}
	defer db.Close()
}

func TestOpenFileBackedDBEscapesReservedPathCharacters(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "ferret #?.db")
	db, err := Open(context.Background(), OpenOptions{Path: stringPtr(path)})
	if err != nil {
		t.Fatalf("unexpected open error: %v", err)
	}
	defer db.Close()

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected database file at original path: %v", err)
	}
}

func TestOpenURIDB(t *testing.T) {
	t.Parallel()

	db, err := Open(context.Background(), OpenOptions{URI: stringPtr("file:ferret_open_test?mode=memory&cache=shared")})
	if err != nil {
		t.Fatalf("unexpected open error: %v", err)
	}
	defer db.Close()
}

func TestOpenValidation(t *testing.T) {
	t.Parallel()

	cases := []struct {
		options OpenOptions
		name    string
		want    string
	}{
		{
			name: "missing source",
			want: "exactly one of path, memory, or uri must be provided",
		},
		{
			name: "memory false is omitted",
			options: OpenOptions{
				Memory: boolPtr(false),
			},
			want: "exactly one of path, memory, or uri must be provided",
		},
		{
			name: "multiple sources",
			options: OpenOptions{
				Memory: boolPtr(true),
				Path:   stringPtr("data.db"),
			},
			want: "exactly one of path, memory, or uri must be provided",
		},
		{
			name: "read only create conflict",
			options: OpenOptions{
				Path:     stringPtr("data.db"),
				ReadOnly: boolPtr(true),
				Create:   boolPtr(true),
			},
			want: "readOnly and create cannot both be true",
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := Open(context.Background(), tt.options)
			assertErrorContains(t, err, tt.want)
		})
	}
}
