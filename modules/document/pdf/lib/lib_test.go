package lib

import (
	"slices"
	"strings"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestRegisterLib(t *testing.T) {
	library := runtime.NewLibrary()
	RegisterLib(library.Namespace("DOCUMENT").Namespace("PDF"))

	funcs, err := library.Build()
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}

	expected := []string{
		"DOCUMENT::PDF::BLOCKS",
		"DOCUMENT::PDF::CLOSE",
		"DOCUMENT::PDF::OPEN",
		"DOCUMENT::PDF::PAGE",
		"DOCUMENT::PDF::PAGES",
		"DOCUMENT::PDF::PAGE_COUNT",
		"DOCUMENT::PDF::PAGE_INFO",
		"DOCUMENT::PDF::TEXT",
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

	for _, name := range names {
		if strings.HasPrefix(name, "PDF::") {
			t.Fatalf("unexpected top-level PDF alias %q", name)
		}
	}
}
