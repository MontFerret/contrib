package lib

import (
	"slices"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestRegisterLib(t *testing.T) {
	library := runtime.NewLibrary()
	if err := RegisterLib(library.Namespace("DB").Namespace("SQLITE")); err != nil {
		t.Fatalf("unexpected registration error: %v", err)
	}

	funcs, err := library.Build()
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}

	expected := []string{
		"DB::SQLITE::BEGIN",
		"DB::SQLITE::CLOSE",
		"DB::SQLITE::COMMIT",
		"DB::SQLITE::OPEN",
		"DB::SQLITE::ROLLBACK",
	}

	if funcs.Size() != len(expected) {
		t.Fatalf("expected %d registered functions, got %d", len(expected), funcs.Size())
	}

	names := funcs.List()
	slices.Sort(names)
	slices.Sort(expected)

	if !slices.Equal(names, expected) {
		t.Fatalf("unexpected registered names: got %v, want %v", names, expected)
	}
}
