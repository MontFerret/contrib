package core

import (
	"context"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestConnectionQueryRowsAndParams(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := openMemoryDB(t, ctx)

	queryExecForTest(t, ctx, db, `CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT NOT NULL, score REAL, note TEXT)`)
	queryExecForTest(t, ctx, db, `INSERT INTO users(name, score, note) VALUES (?, ?, ?)`,
		runtime.NewString("Ada"),
		runtime.NewFloat(9.5),
		runtime.None,
	)

	rows := queryRowsForTest(t, ctx, db, `SELECT id, name, score, note FROM users WHERE name = ?`, runtime.NewString("Ada"))
	assertArrayLen(t, ctx, rows, 1)

	row := mustObjectAt(t, ctx, rows, 0)
	if got := objectField(t, ctx, row, "id"); got != runtime.NewInt64(1) {
		t.Fatalf("expected id 1, got %v", got)
	}
	if got := objectField(t, ctx, row, "name"); got != runtime.NewString("Ada") {
		t.Fatalf("expected name Ada, got %v", got)
	}
	if got := objectField(t, ctx, row, "score"); got != runtime.NewFloat(9.5) {
		t.Fatalf("expected score 9.5, got %v", got)
	}
	if got := objectField(t, ctx, row, "note"); got != runtime.None {
		t.Fatalf("expected note None, got %v", got)
	}
}

func TestConnectionQueryUsesParamsIndependentlyFromOptions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := openMemoryDB(t, ctx)

	out, err := db.Query(ctx, runtime.Query{
		Kind:       runtime.NewString("sql"),
		Expression: runtime.NewString(`SELECT ? AS value`),
		Params: runtime.NewObjectWith(map[string]runtime.Value{
			"params": runtime.NewArrayWith(runtime.NewString("from-params")),
		}),
		Options: runtime.NewString("ignored-options"),
	})
	if err != nil {
		t.Fatalf("unexpected query error: %v", err)
	}

	rows, ok := out.(*runtime.Array)
	if !ok {
		t.Fatalf("expected runtime array, got %T", out)
	}

	row := mustObjectAt(t, ctx, rows, 0)
	if got := objectField(t, ctx, row, "value"); got != runtime.NewString("from-params") {
		t.Fatalf("expected query parameter value, got %v", got)
	}
}

func TestConnectionQueryTypeMapping(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := openMemoryDB(t, ctx)

	rows := queryRowsForTest(t, ctx, db, `SELECT ? AS i, ? AS f, ? AS s, ? AS b, ? AS blob, NULL AS n`,
		runtime.NewInt(42),
		runtime.NewFloat(4.25),
		runtime.NewString("ferret"),
		runtime.True,
		runtime.NewBinary([]byte("bin")),
	)
	assertArrayLen(t, ctx, rows, 1)

	row := mustObjectAt(t, ctx, rows, 0)
	if got := objectField(t, ctx, row, "i"); got != runtime.NewInt64(42) {
		t.Fatalf("expected int 42, got %v", got)
	}
	if got := objectField(t, ctx, row, "f"); got != runtime.NewFloat(4.25) {
		t.Fatalf("expected float 4.25, got %v", got)
	}
	if got := objectField(t, ctx, row, "s"); got != runtime.NewString("ferret") {
		t.Fatalf("expected string ferret, got %v", got)
	}
	if got := objectField(t, ctx, row, "b"); got != runtime.NewInt(1) {
		t.Fatalf("expected bool param to decode as SQLite integer 1, got %v", got)
	}
	if got := objectField(t, ctx, row, "blob"); runtime.CompareValues(got, runtime.NewBinary([]byte("bin"))) != 0 {
		t.Fatalf("expected binary value, got %v", got)
	}
	if got := objectField(t, ctx, row, "n"); got != runtime.None {
		t.Fatalf("expected none, got %v", got)
	}
}

func TestQueryModifiers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := openMemoryDB(t, ctx)

	queryExecForTest(t, ctx, db, `CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT NOT NULL)`)
	queryExecForTest(t, ctx, db, `INSERT INTO users(name) VALUES (?)`, runtime.NewString("Ada"))
	queryExecForTest(t, ctx, db, `INSERT INTO users(name) VALUES (?)`, runtime.NewString("Grace"))

	query := runtime.Query{
		Kind:       runtime.NewString("sql"),
		Expression: runtime.NewString(`SELECT id, name FROM users ORDER BY id`),
	}

	one, err := db.QueryOne(ctx, query)
	if err != nil {
		t.Fatalf("unexpected query one error: %v", err)
	}
	first, ok := one.(*runtime.Object)
	if !ok {
		t.Fatalf("expected object, got %T", one)
	}
	if got := objectField(t, ctx, first, "name"); got != runtime.NewString("Ada") {
		t.Fatalf("expected first name Ada, got %v", got)
	}

	count, err := db.QueryCount(ctx, query)
	if err != nil {
		t.Fatalf("unexpected query count error: %v", err)
	}
	if count != runtime.NewInt(2) {
		t.Fatalf("expected count 2, got %v", count)
	}

	exists, err := db.QueryExists(ctx, query)
	if err != nil {
		t.Fatalf("unexpected query exists error: %v", err)
	}
	if exists != runtime.True {
		t.Fatalf("expected exists true, got %v", exists)
	}

	missing, err := db.QueryOne(ctx, runtime.Query{
		Kind:       runtime.NewString("sql"),
		Expression: runtime.NewString(`SELECT id FROM users WHERE id = 99`),
	})
	if err != nil {
		t.Fatalf("unexpected missing query one error: %v", err)
	}
	if missing != runtime.None {
		t.Fatalf("expected missing query to return none, got %v", missing)
	}
}

func TestConnectionQueryExecMetadataAndModifiers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := openMemoryDB(t, ctx)

	createList, err := db.Query(ctx, runtime.Query{
		Kind:       runtime.NewString("SQL_EXEC"),
		Expression: runtime.NewString(`CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT NOT NULL)`),
	})
	if err != nil {
		t.Fatalf("unexpected query exec error: %v", err)
	}
	createRows, ok := createList.(*runtime.Array)
	if !ok {
		t.Fatalf("expected array, got %T", createList)
	}
	assertArrayLen(t, ctx, createRows, 1)
	assertExecMetadata(t, ctx, mustObjectAt(t, ctx, createRows, 0), runtime.NewInt64(0), runtime.None)

	insert := queryExecForTest(t, ctx, db, `INSERT INTO users(name) VALUES (?)`, runtime.NewString("Ada"))
	assertExecMetadata(t, ctx, insert, runtime.NewInt64(1), runtime.NewInt64(1))

	cteInsert := queryExecForTest(
		t,
		ctx,
		db,
		`WITH incoming(name) AS (SELECT ?) INSERT INTO users(name) SELECT name FROM incoming`,
		runtime.NewString("Lovelace"),
	)
	assertExecMetadata(t, ctx, cteInsert, runtime.NewInt64(1), runtime.NewInt64(2))

	update := queryExecForTest(
		t,
		ctx,
		db,
		`UPDATE users SET name = ? WHERE id = ?`,
		runtime.NewString("Grace"),
		runtime.NewInt(1),
	)
	assertExecMetadata(t, ctx, update, runtime.NewInt64(1), runtime.None)

	cteUpdate := queryExecForTest(
		t,
		ctx,
		db,
		`WITH selected(id) AS (SELECT ?) UPDATE users SET name = ? WHERE id = (SELECT id FROM selected)`,
		runtime.NewInt(2),
		runtime.NewString("Byron"),
	)
	assertExecMetadata(t, ctx, cteUpdate, runtime.NewInt64(1), runtime.None)

	count, err := db.QueryCount(ctx, runtime.Query{
		Kind:       runtime.NewString("sql_exec"),
		Expression: runtime.NewString(`UPDATE users SET name = ? WHERE id = ?`),
		Params: runtime.NewObjectWith(map[string]runtime.Value{
			"params": runtime.NewArrayWith(runtime.NewString("Missing"), runtime.NewInt(99)),
		}),
	})
	if err != nil {
		t.Fatalf("unexpected query count exec error: %v", err)
	}
	if count != runtime.NewInt(1) {
		t.Fatalf("expected query count exec to return 1, got %v", count)
	}

	exists, err := db.QueryExists(ctx, runtime.Query{
		Kind:       runtime.NewString("sql_exec"),
		Expression: runtime.NewString(`DELETE FROM users WHERE id = ?`),
		Params: runtime.NewObjectWith(map[string]runtime.Value{
			"params": runtime.NewArrayWith(runtime.NewInt(99)),
		}),
	})
	if err != nil {
		t.Fatalf("unexpected query exists exec error: %v", err)
	}
	if exists != runtime.True {
		t.Fatalf("expected query exists exec to return true, got %v", exists)
	}

	rows := queryRowsForTest(t, ctx, db, `SELECT name FROM users WHERE id = ?`, runtime.NewInt(1))
	assertArrayLen(t, ctx, rows, 1)
	row := mustObjectAt(t, ctx, rows, 0)
	if got := objectField(t, ctx, row, "name"); got != runtime.NewString("Grace") {
		t.Fatalf("expected name Grace, got %v", got)
	}

	rows = queryRowsForTest(t, ctx, db, `SELECT name FROM users WHERE id = ?`, runtime.NewInt(2))
	assertArrayLen(t, ctx, rows, 1)
	row = mustObjectAt(t, ctx, rows, 0)
	if got := objectField(t, ctx, row, "name"); got != runtime.NewString("Byron") {
		t.Fatalf("expected name Byron, got %v", got)
	}
}

func TestTransactionQueryExecMetadata(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := openMemoryDB(t, ctx)
	queryExecForTest(t, ctx, db, `CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT NOT NULL)`)

	committed, err := db.Begin(ctx)
	if err != nil {
		t.Fatalf("unexpected begin error: %v", err)
	}
	insert := queryExecForTest(t, ctx, committed, `INSERT INTO users(name) VALUES (?)`, runtime.NewString("Ada"))
	assertExecMetadata(t, ctx, insert, runtime.NewInt64(1), runtime.NewInt64(1))
	if err := committed.Commit(); err != nil {
		t.Fatalf("unexpected commit error: %v", err)
	}

	rolledBack, err := db.Begin(ctx)
	if err != nil {
		t.Fatalf("unexpected begin error: %v", err)
	}
	queryExecForTest(t, ctx, rolledBack, `INSERT INTO users(name) VALUES (?)`, runtime.NewString("Grace"))
	if err := rolledBack.Rollback(); err != nil {
		t.Fatalf("unexpected rollback error: %v", err)
	}

	rows := queryRowsForTest(t, ctx, db, `SELECT name FROM users ORDER BY id`)
	assertArrayLen(t, ctx, rows, 1)
	row := mustObjectAt(t, ctx, rows, 0)
	if got := objectField(t, ctx, row, "name"); got != runtime.NewString("Ada") {
		t.Fatalf("expected name Ada, got %v", got)
	}
}

func TestQueryErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := openMemoryDB(t, ctx)

	_, err := db.Query(ctx, runtime.Query{Kind: runtime.NewString("css"), Expression: runtime.NewString("body")})
	assertErrorContains(t, err, `unsupported dialect "css"; expected "sql" or "sql_exec"`)

	if err := db.Close(); err != nil {
		t.Fatalf("unexpected close error: %v", err)
	}

	_, err = db.Query(ctx, runtime.Query{Kind: runtime.NewString("sql"), Expression: runtime.NewString("SELECT 1")})
	assertErrorContains(t, err, "database connection has been closed")
}
