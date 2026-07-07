package lib

import (
	"slices"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestRegisterLib(t *testing.T) {
	library := runtime.NewLibrary()
	RegisterLib(library.Namespace("DOCUMENT").Namespace("XLSX"))

	funcs, err := library.Build()
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}

	expected := []string{
		"DOCUMENT::XLSX::ADD_SHEET",
		"DOCUMENT::XLSX::APPEND",
		"DOCUMENT::XLSX::CLOSE",
		"DOCUMENT::XLSX::CREATE",
		"DOCUMENT::XLSX::DELETE_SHEET",
		"DOCUMENT::XLSX::GET",
		"DOCUMENT::XLSX::OPEN",
		"DOCUMENT::XLSX::RANGE",
		"DOCUMENT::XLSX::SAVE",
		"DOCUMENT::XLSX::SAVE_AS",
		"DOCUMENT::XLSX::SET",
		"DOCUMENT::XLSX::SHEET",
		"DOCUMENT::XLSX::SHEETS",
		"DOCUMENT::XLSX::WRITE_RANGE",
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
