package lib

import (
	"slices"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestRegisterLib(t *testing.T) {
	library := runtime.NewLibrary()
	if err := RegisterLib(library.Namespace("WEB").Namespace("SITEMAP")); err != nil {
		t.Fatalf("unexpected registration error: %v", err)
	}

	funcs, err := library.Build()
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}

	expected := []string{
		"WEB::SITEMAP::FETCH",
		"WEB::SITEMAP::STREAM",
		"WEB::SITEMAP::URLS",
	}

	if funcs.Size() != len(expected) {
		t.Fatalf("expected %d registered functions, got %d", len(expected), funcs.Size())
	}

	names := funcs.List()
	slices.Sort(names)
	slices.Sort(expected)

	if !slices.Equal(names, expected) {
		t.Fatalf("unexpected names: got %v, want %v", names, expected)
	}
}
