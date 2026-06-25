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
	RegisterLib(library.Namespace("DB").Namespace("POSTGRES"))

	funcs, err := library.Build()
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}

	expected := []string{
		"DB::POSTGRES::BEGIN",
		"DB::POSTGRES::CLOSE",
		"DB::POSTGRES::COMMIT",
		"DB::POSTGRES::OPEN",
		"DB::POSTGRES::ROLLBACK",
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

func TestLifecycleFunctionsRejectWrongHandles(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	bad := runtime.NewString("invalid")

	_, err := Close(ctx, bad)
	assertErrorContains(t, err, "expected Postgres database handle")

	_, err = Begin(ctx, bad)
	assertErrorContains(t, err, "expected Postgres database handle")

	_, err = Commit(ctx, bad)
	assertErrorContains(t, err, "expected Postgres transaction handle")

	_, err = Rollback(ctx, bad)
	assertErrorContains(t, err, "expected Postgres transaction handle")
}

func assertErrorContains(t *testing.T, err error, expected string) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected error containing %q", expected)
	}
	if !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected error containing %q, got %q", expected, err.Error())
	}
}
