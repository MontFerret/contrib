package lib

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/MontFerret/contrib/modules/archive/core"
	ferretfs "github.com/MontFerret/ferret/v2/pkg/fs"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestRegisterLib(t *testing.T) {
	library := runtime.NewLibrary()
	if err := RegisterLib(library.Namespace("ARCHIVE")); err != nil {
		t.Fatalf("unexpected registration error: %v", err)
	}

	functions, err := library.Build()
	if err != nil {
		t.Fatalf("unexpected library build error: %v", err)
	}

	expected := []string{
		"ARCHIVE::EXTRACT",
		"ARCHIVE::LIST",
		"ARCHIVE::READ",
	}
	actual := functions.List()
	slices.Sort(actual)

	if functions.Size() != len(expected) {
		t.Fatalf("registered %d functions, want %d", functions.Size(), len(expected))
	}
	if !slices.Equal(actual, expected) {
		t.Fatalf("registered names = %v, want %v", actual, expected)
	}
}

func TestHandlersRequireConfigBeforeFilesystem(t *testing.T) {
	tests := []struct {
		call func(context.Context) (runtime.Value, error)
		name string
	}{
		{
			name: "list",
			call: func(ctx context.Context) (runtime.Value, error) {
				return List(ctx, runtime.NewString("source.zip"))
			},
		},
		{
			name: "read",
			call: func(ctx context.Context) (runtime.Value, error) {
				return Read(ctx, runtime.NewString("source.zip"), runtime.NewString("entry.txt"))
			},
		},
		{
			name: "extract",
			call: func(ctx context.Context) (runtime.Value, error) {
				return Extract(ctx, runtime.NewString("source.zip"), runtime.NewString("destination"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := test.call(context.Background())
			if !errors.Is(err, core.ErrConfigNotFound) {
				t.Fatalf("expected ErrConfigNotFound, got %v", err)
			}
			if errors.Is(err, ferretfs.ErrNotFound) {
				t.Fatalf("filesystem resolution happened before config resolution: %v", err)
			}
		})
	}
}

func TestHandlersResolveFilesystemAfterConfig(t *testing.T) {
	ctx := core.WithConfig(context.Background(), core.DefaultConfig())
	tests := []struct {
		call func(context.Context) (runtime.Value, error)
		name string
	}{
		{
			name: "list",
			call: func(ctx context.Context) (runtime.Value, error) {
				return List(ctx, runtime.NewString("source.zip"))
			},
		},
		{
			name: "read",
			call: func(ctx context.Context) (runtime.Value, error) {
				return Read(ctx, runtime.NewString("source.zip"), runtime.NewString("entry.txt"))
			},
		},
		{
			name: "extract",
			call: func(ctx context.Context) (runtime.Value, error) {
				return Extract(ctx, runtime.NewString("source.zip"), runtime.NewString("destination"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := test.call(ctx)
			if !errors.Is(err, ferretfs.ErrNotFound) {
				t.Fatalf("expected filesystem ErrNotFound, got %v", err)
			}
			if errors.Is(err, core.ErrConfigNotFound) {
				t.Fatalf("configuration was not resolved: %v", err)
			}
		})
	}
}

func TestHandlersValidateArgumentsBeforeConfig(t *testing.T) {
	unknownOptions := runtime.NewObjectWith(map[string]runtime.Value{
		"unknown": runtime.True,
	})
	tests := []struct {
		call func() (runtime.Value, error)
		name string
	}{
		{
			name: "list arity",
			call: func() (runtime.Value, error) {
				return List(context.Background())
			},
		},
		{
			name: "list source type",
			call: func() (runtime.Value, error) {
				return List(context.Background(), runtime.NewInt(1))
			},
		},
		{
			name: "read arity",
			call: func() (runtime.Value, error) {
				return Read(context.Background(), runtime.NewString("source.zip"))
			},
		},
		{
			name: "read name type",
			call: func() (runtime.Value, error) {
				return Read(context.Background(), runtime.NewString("source.zip"), runtime.NewInt(1))
			},
		},
		{
			name: "extract arity",
			call: func() (runtime.Value, error) {
				return Extract(context.Background(), runtime.NewString("source.zip"))
			},
		},
		{
			name: "extract options",
			call: func() (runtime.Value, error) {
				return Extract(
					context.Background(),
					runtime.NewString("source.zip"),
					runtime.NewString("destination"),
					unknownOptions,
				)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := test.call()
			if err == nil {
				t.Fatal("expected argument error")
			}
			if errors.Is(err, core.ErrConfigNotFound) {
				t.Fatalf("configuration was resolved before argument validation: %v", err)
			}
			if errors.Is(err, ferretfs.ErrNotFound) {
				t.Fatalf("filesystem was resolved before argument validation: %v", err)
			}
		})
	}
}
