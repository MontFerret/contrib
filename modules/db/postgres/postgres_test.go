package postgres

import (
	"context"
	"strings"
	"testing"

	"github.com/MontFerret/ferret/v2"
	"github.com/MontFerret/ferret/v2/pkg/source"
)

func TestNewSmoke(t *testing.T) {
	mod := New()

	if mod == nil {
		t.Fatal("expected module to be non-nil")
	}

	if mod.Name() != "db/postgres" {
		t.Fatalf("expected module name %q, got %q", "db/postgres", mod.Name())
	}
}

func TestOpenValidationThroughFerret(t *testing.T) {
	t.Parallel()

	_, err := runFQL(t, `RETURN DB::POSTGRES::OPEN({})`)
	assertErrorContains(t, err, "exactly one of uri or structured connection fields must be provided")
}

func TestCloseRejectsWrongHandleThroughFerret(t *testing.T) {
	t.Parallel()

	_, err := runFQL(t, `RETURN DB::POSTGRES::CLOSE("invalid")`)
	assertErrorContains(t, err, "expected Postgres database handle")
}

func runFQL(t *testing.T, query string) (*ferret.Output, error) {
	t.Helper()

	engine, err := ferret.New(ferret.WithModules(New()))
	if err != nil {
		t.Fatalf("unexpected engine error: %v", err)
	}
	t.Cleanup(func() {
		if err := engine.Close(); err != nil {
			t.Fatalf("unexpected engine close error: %v", err)
		}
	})

	return engine.Run(context.Background(), source.NewAnonymous(query))
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
