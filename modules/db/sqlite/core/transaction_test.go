package core

import (
	"context"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestTransactionCommitPersistsChanges(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := openMemoryDB(t, ctx)
	queryExecForTest(t, ctx, db, `CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT NOT NULL)`)

	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatalf("unexpected begin error: %v", err)
	}

	queryExecForTest(t, ctx, tx, `INSERT INTO users(name) VALUES (?)`, runtime.NewString("Ada"))

	inTx := queryRowsForTest(t, ctx, tx, `SELECT name FROM users`)
	assertArrayLen(t, ctx, inTx, 1)

	if err := tx.Commit(); err != nil {
		t.Fatalf("unexpected commit error: %v", err)
	}

	rows := queryRowsForTest(t, ctx, db, `SELECT name FROM users`)
	assertArrayLen(t, ctx, rows, 1)
}

func TestTransactionRollbackDiscardsChanges(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := openMemoryDB(t, ctx)
	queryExecForTest(t, ctx, db, `CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT NOT NULL)`)

	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatalf("unexpected begin error: %v", err)
	}

	queryExecForTest(t, ctx, tx, `INSERT INTO users(name) VALUES (?)`, runtime.NewString("Ada"))

	if err := tx.Rollback(); err != nil {
		t.Fatalf("unexpected rollback error: %v", err)
	}

	rows := queryRowsForTest(t, ctx, db, `SELECT name FROM users`)
	assertArrayLen(t, ctx, rows, 0)
}

func TestTransactionAfterFinishFails(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := openMemoryDB(t, ctx)
	queryExecForTest(t, ctx, db, `CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT NOT NULL)`)

	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatalf("unexpected begin error: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("unexpected commit error: %v", err)
	}

	_, err = tx.Query(ctx, runtime.Query{Kind: runtime.NewString("sql"), Payload: runtime.NewString("SELECT 1")})
	assertErrorContains(t, err, "transaction has already been finished")
}

func TestActiveTransactionRollsBackOnClose(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := openMemoryDB(t, ctx)
	queryExecForTest(t, ctx, db, `CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT NOT NULL)`)

	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatalf("unexpected begin error: %v", err)
	}
	queryExecForTest(t, ctx, tx, `INSERT INTO users(name) VALUES (?)`, runtime.NewString("Ada"))

	if err := tx.Close(); err != nil {
		t.Fatalf("unexpected tx close error: %v", err)
	}

	rows := queryRowsForTest(t, ctx, db, `SELECT name FROM users`)
	assertArrayLen(t, ctx, rows, 0)
}

func TestClosingDBInvalidatesTransaction(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := openMemoryDB(t, ctx)
	queryExecForTest(t, ctx, db, `CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT NOT NULL)`)

	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatalf("unexpected begin error: %v", err)
	}

	if err := db.Close(); err != nil {
		t.Fatalf("unexpected db close error: %v", err)
	}

	_, err = tx.Query(ctx, runtime.Query{Kind: runtime.NewString("sql"), Payload: runtime.NewString("SELECT 1")})
	assertErrorContains(t, err, "parent database has been closed")
}
