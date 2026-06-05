package sqlite

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/MontFerret/ferret/v2"
	"github.com/MontFerret/ferret/v2/pkg/module"
	"github.com/MontFerret/ferret/v2/pkg/source"
)

func TestNewSmoke(t *testing.T) {
	mod := New()

	if mod == nil {
		t.Fatal("expected module to be non-nil")
	}

	if mod.Name() != "db/sqlite" {
		t.Fatalf("expected module name %q, got %q", "db/sqlite", mod.Name())
	}
}

func TestDefaultModuleAllowsFileDB(t *testing.T) {
	t.Parallel()

	path := t.TempDir() + "/ferret.db"
	_, err := runFQL(t, New(), fmt.Sprintf(`
		LET db = DB::SQLITE::OPEN({ path: %q })
		RETURN DB::SQLITE::CLOSE(db)
	`, path))
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
}

func TestMemoryOnlyModuleRejectsFileDB(t *testing.T) {
	t.Parallel()

	path := t.TempDir() + "/ferret.db"
	_, err := runFQL(t, New(WithMemoryOnly()), fmt.Sprintf(`
		LET db = DB::SQLITE::OPEN({ path: %q })
		RETURN DB::SQLITE::CLOSE(db)
	`, path))
	if err == nil {
		t.Fatal("expected file-backed open to fail")
	}
	if !strings.Contains(err.Error(), "file-backed SQLite databases are disabled") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMemoryOnlyModuleAllowsMemoryDB(t *testing.T) {
	t.Parallel()

	_, err := runFQL(t, New(WithMemoryOnly()), `
		LET db = DB::SQLITE::OPEN({ memory: true })
		RETURN DB::SQLITE::CLOSE(db)
	`)
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
}

func runFQL(t *testing.T, mod module.Module, query string) (*ferret.Output, error) {
	t.Helper()

	engine, err := ferret.New(ferret.WithModules(mod))
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
