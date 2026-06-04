package core

import (
	"context"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestDispatchMetadata(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := openMemoryDB(t, ctx)

	create := dispatchSQLForTest(t, ctx, db, `CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT NOT NULL)`)
	createObj := create.(*runtime.Object)
	if got := objectField(t, ctx, createObj, "rowsAffected"); got != runtime.NewInt64(0) {
		t.Fatalf("expected create rowsAffected 0, got %v", got)
	}
	if got := objectField(t, ctx, createObj, "lastInsertId"); got != runtime.None {
		t.Fatalf("expected create lastInsertId none, got %v", got)
	}

	insert := dispatchSQLForTest(t, ctx, db, `INSERT INTO users(name) VALUES (?)`, runtime.NewString("Ada"))
	insertObj := insert.(*runtime.Object)
	if got := objectField(t, ctx, insertObj, "rowsAffected"); got != runtime.NewInt64(1) {
		t.Fatalf("expected insert rowsAffected 1, got %v", got)
	}
	if got := objectField(t, ctx, insertObj, "lastInsertId"); got != runtime.NewInt64(1) {
		t.Fatalf("expected insert lastInsertId 1, got %v", got)
	}

	update := dispatchSQLForTest(t, ctx, db, `UPDATE users SET name = ? WHERE id = ?`, runtime.NewString("Grace"), runtime.NewInt(1))
	updateObj := update.(*runtime.Object)
	if got := objectField(t, ctx, updateObj, "rowsAffected"); got != runtime.NewInt64(1) {
		t.Fatalf("expected update rowsAffected 1, got %v", got)
	}
	if got := objectField(t, ctx, updateObj, "lastInsertId"); got != runtime.None {
		t.Fatalf("expected update lastInsertId none, got %v", got)
	}
}

func TestDispatchErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := openMemoryDB(t, ctx)

	_, err := db.Dispatch(ctx, runtime.DispatchEvent{
		Name:    runtime.NewString("css"),
		Payload: runtime.NewString("body"),
	})
	assertErrorContains(t, err, `unsupported dialect "css"; expected "sql"`)

	_, err = db.Dispatch(ctx, runtime.DispatchEvent{
		Name:    runtime.NewString("sql"),
		Payload: runtime.NewString("SELECT ?"),
		Options: runtime.NewObjectWith(map[string]runtime.Value{
			"params": runtime.NewArrayWith(runtime.NewArray(0)),
		}),
	})
	assertErrorContains(t, err, "unsupported param type Array")

	if err := db.Close(); err != nil {
		t.Fatalf("unexpected close error: %v", err)
	}

	_, err = db.Dispatch(ctx, runtime.DispatchEvent{
		Name:    runtime.NewString("sql"),
		Payload: runtime.NewString("CREATE TABLE users(id INTEGER)"),
	})
	assertErrorContains(t, err, "database connection has been closed")
}
