package lib

import (
	"slices"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestRegisterLib(t *testing.T) {
	library := runtime.NewLibrary()
	RegisterLib(library.Namespace("XML"))

	funcs, err := library.Build()
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}

	expected := []string{
		"XML::ATTR",
		"XML::CHILDREN",
		"XML::DECODE",
		"XML::DECODE_STREAM",
		"XML::ENCODE",
		"XML::ROOT",
		"XML::TEXT",
	}

	if funcs.Size() != len(expected) {
		t.Fatalf("expected %d registered functions, got %d", len(expected), funcs.Size())
	}

	for _, name := range expected {
		if !funcs.Has(name) {
			t.Fatalf("expected function %q to be registered", name)
		}
	}

	names := funcs.List()
	slices.Sort(names)
	slices.Sort(expected)

	if !slices.Equal(names, expected) {
		t.Fatalf("unexpected registered names: got %v, want %v", names, expected)
	}
}
