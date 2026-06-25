package core

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestIntegrationPostgresLifecycle(t *testing.T) {
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		t.Skip("POSTGRES_DSN is not set")
	}

	ctx := context.Background()
	db, err := Open(ctx, OpenOptions{URI: &dsn})
	if err != nil {
		t.Fatalf("unexpected open error: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("unexpected close error: %v", err)
		}
	})

	table := fmt.Sprintf("ferret_postgres_%d", time.Now().UnixNano())
	queryExecForTest(t, ctx, db, fmt.Sprintf("CREATE TABLE %s (id BIGSERIAL PRIMARY KEY, name TEXT NOT NULL)", table))
	t.Cleanup(func() {
		_, _ = db.Query(ctx, runtime.Query{
			Kind:       runtime.NewString("sql_exec"),
			Expression: runtime.NewString(fmt.Sprintf("DROP TABLE IF EXISTS %s", table)),
		})
	})

	inserted := queryRowsForTest(
		t,
		ctx,
		db,
		fmt.Sprintf("INSERT INTO %s(name) VALUES ($1) RETURNING id, name", table),
		runtime.NewString("Ada"),
	)
	assertArrayLen(t, ctx, inserted, 1)
	insertedRow := mustObjectAt(t, ctx, inserted, 0)
	if got := objectField(t, ctx, insertedRow, "name"); got != runtime.NewString("Ada") {
		t.Fatalf("expected inserted name Ada, got %v", got)
	}

	updated := queryExecForTest(
		t,
		ctx,
		db,
		fmt.Sprintf("UPDATE %s SET name = $1 WHERE name = $2", table),
		runtime.NewString("Grace"),
		runtime.NewString("Ada"),
	)
	assertExecMetadata(t, ctx, updated, runtime.NewInt64(1), runtime.None)

	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatalf("unexpected begin error: %v", err)
	}
	queryExecForTest(t, ctx, tx, fmt.Sprintf("INSERT INTO %s(name) VALUES ($1)", table), runtime.NewString("Rolled Back"))
	if err := tx.Rollback(); err != nil {
		t.Fatalf("unexpected rollback error: %v", err)
	}

	rows := queryRowsForTest(t, ctx, db, fmt.Sprintf("SELECT name FROM %s ORDER BY id", table))
	assertArrayLen(t, ctx, rows, 1)
	row := mustObjectAt(t, ctx, rows, 0)
	if got := objectField(t, ctx, row, "name"); got != runtime.NewString("Grace") {
		t.Fatalf("expected final name Grace, got %v", got)
	}
}
