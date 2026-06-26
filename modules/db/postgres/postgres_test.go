package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/MontFerret/ferret/v2"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
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

func TestIntegrationPostgresThroughFerret(t *testing.T) {
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		t.Skip("POSTGRES_DSN is not set")
	}

	table := fmt.Sprintf("ferret_postgres_fql_%d", time.Now().UnixNano())
	initialDropSQL := rawFQLString(fmt.Sprintf("DROP TABLE IF EXISTS %s", table))
	createSQL := rawFQLString(fmt.Sprintf("CREATE TABLE %s (id BIGSERIAL PRIMARY KEY, name TEXT NOT NULL)", table))
	insertSQL := rawFQLString(fmt.Sprintf("INSERT INTO %s(name) VALUES ($1) RETURNING id, name", table))
	updateSQL := rawFQLString(fmt.Sprintf("UPDATE %s SET name = $1 WHERE name = $2", table))
	rollbackInsertSQL := rawFQLString(fmt.Sprintf("INSERT INTO %s(name) VALUES ($1)", table))
	selectSQL := rawFQLString(fmt.Sprintf("SELECT name FROM %s ORDER BY id", table))
	finalDropSQL := rawFQLString(fmt.Sprintf("DROP TABLE IF EXISTS %s", table))

	output, err := runFQL(t, fmt.Sprintf(`
		LET db = DB::POSTGRES::OPEN({ uri: @dsn })
		LET initialDrop = QUERY ONE %s IN db USING sql_exec
		LET created = QUERY ONE %s IN db USING sql_exec
		LET inserted = QUERY ONE %s IN db USING sql WITH {
			params: ["Ada"]
		}
		LET updated = QUERY ONE %s IN db USING sql_exec WITH {
			params: ["Grace", "Ada"]
		}
		LET tx = DB::POSTGRES::BEGIN(db)
		LET rollbackInsert = QUERY ONE %s IN tx USING sql_exec WITH {
			params: ["Rolled Back"]
		}
		LET rollback = DB::POSTGRES::ROLLBACK(tx)
		LET rows = QUERY %s IN db USING sql
		LET finalDrop = QUERY ONE %s IN db USING sql_exec
		LET closed = DB::POSTGRES::CLOSE(db)
		RETURN inserted.name == "Ada"
			AND updated.rowsAffected == 1
			AND updated.lastInsertId == NONE
			AND LENGTH(rows) == 1
			AND rows[0].name == "Grace"
	`, initialDropSQL, createSQL, insertSQL, updateSQL, rollbackInsertSQL, selectSQL, finalDropSQL),
		ferret.WithRuntimeParam("dsn", runtime.NewString(dsn)),
	)
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}

	assertOutputBool(t, output, true)
}

func runFQL(t *testing.T, query string, opts ...ferret.Option) (*ferret.Output, error) {
	t.Helper()

	engineOpts := append([]ferret.Option{ferret.WithModules(New())}, opts...)
	engine, err := ferret.New(engineOpts...)
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

func rawFQLString(value string) string {
	return "`" + value + "`"
}

func assertOutputBool(t *testing.T, output *ferret.Output, expected bool) {
	t.Helper()

	var actual bool
	if err := json.Unmarshal(output.Content, &actual); err != nil {
		t.Fatalf("failed to decode output bool: %v", err)
	}
	if actual != expected {
		t.Fatalf("expected output %v, got %v", expected, actual)
	}
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
