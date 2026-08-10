package lib

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestRegisterLib(t *testing.T) {
	library := runtime.NewLibrary()
	if err := RegisterLib(library.Namespace("DB").Namespace("REDIS")); err != nil {
		t.Fatalf("unexpected registration error: %v", err)
	}

	functions, err := library.Build()
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}

	expected := []string{
		"DB::REDIS::CLOSE",
		"DB::REDIS::OPEN",
	}
	actual := functions.List()
	slices.Sort(actual)
	slices.Sort(expected)

	if !slices.Equal(actual, expected) {
		t.Fatalf("unexpected registered names: got %v, want %v", actual, expected)
	}
}

func TestCloseRejectsWrongHandle(t *testing.T) {
	t.Parallel()

	_, err := Close(context.Background(), runtime.NewString("invalid"))
	if err == nil || !strings.Contains(err.Error(), "expected Redis connection handle") {
		t.Fatalf("unexpected error: %v", err)
	}
}
