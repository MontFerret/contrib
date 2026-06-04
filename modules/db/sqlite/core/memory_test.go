package core

import (
	"context"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestPrivateMemoryDatabasesAreIsolated(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	first := openMemoryDB(t, ctx)
	second := openMemoryDB(t, ctx)

	dispatchSQLForTest(t, ctx, first, `CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT NOT NULL)`)
	dispatchSQLForTest(t, ctx, first, `INSERT INTO users(name) VALUES (?)`, runtime.NewString("Ada"))

	_, err := second.Query(ctx, runtime.Query{Kind: runtime.NewString("sql"), Payload: runtime.NewString(`SELECT name FROM users`)})
	assertErrorContains(t, err, "no such table: users")
}

func TestSharedURIMemoryDatabase(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	uri := "file:ferret_shared_uri_unit?mode=memory&cache=shared"
	first, err := Open(ctx, OpenOptions{URI: stringPtr(uri)})
	if err != nil {
		t.Fatalf("unexpected first open error: %v", err)
	}
	defer first.Close()

	second, err := Open(ctx, OpenOptions{URI: stringPtr(uri)})
	if err != nil {
		t.Fatalf("unexpected second open error: %v", err)
	}
	defer second.Close()

	dispatchSQLForTest(t, ctx, first, `CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT NOT NULL)`)
	dispatchSQLForTest(t, ctx, first, `INSERT INTO users(name) VALUES (?)`, runtime.NewString("Ada"))

	rows := queryRowsForTest(t, ctx, second, `SELECT name FROM users`)
	assertArrayLen(t, ctx, rows, 1)
}
