package core

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestConnectionQueryRowsAndParams(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, mock := mockDB(t)
	conn := NewConnection(db)
	now := time.Date(2026, 6, 25, 10, 30, 0, 0, time.UTC)

	mock.ExpectQuery("SELECT id, name, active, created_at, data, note FROM users WHERE id = $1").
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "active", "created_at", "data", "note"}).
			AddRow(int64(1), "Ada", true, now, []byte("bin"), nil))

	rows := queryRowsForTest(
		t,
		ctx,
		conn,
		"SELECT id, name, active, created_at, data, note FROM users WHERE id = $1",
		runtime.NewInt(1),
	)
	assertArrayLen(t, ctx, rows, 1)

	row := mustObjectAt(t, ctx, rows, 0)
	if got := objectField(t, ctx, row, "id"); got != runtime.NewInt64(1) {
		t.Fatalf("expected id 1, got %v", got)
	}
	if got := objectField(t, ctx, row, "name"); got != runtime.NewString("Ada") {
		t.Fatalf("expected name Ada, got %v", got)
	}
	if got := objectField(t, ctx, row, "active"); got != runtime.True {
		t.Fatalf("expected active true, got %v", got)
	}
	if got := objectField(t, ctx, row, "created_at"); got != runtime.NewDateTime(now) {
		t.Fatalf("expected datetime %v, got %v", now, got)
	}
	if got := objectField(t, ctx, row, "data"); runtime.CompareValues(got, runtime.NewBinary([]byte("bin"))) != 0 {
		t.Fatalf("expected binary value, got %v", got)
	}
	if got := objectField(t, ctx, row, "note"); got != runtime.None {
		t.Fatalf("expected note None, got %v", got)
	}
}

func TestConnectionQueryUsesParamsIndependentlyFromOptions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, mock := mockDB(t)
	conn := NewConnection(db)

	mock.ExpectQuery("SELECT $1 AS value").
		WithArgs("from-params").
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("from-params"))

	out, err := conn.Query(ctx, runtime.Query{
		Kind:       runtime.NewString("sql"),
		Expression: runtime.NewString("SELECT $1 AS value"),
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

func TestQueryModifiers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, mock := mockDB(t)
	conn := NewConnection(db)
	sqlText := "SELECT id, name FROM users ORDER BY id"

	for i := 0; i < 3; i++ {
		mock.ExpectQuery(sqlText).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(int64(1), "Ada").
				AddRow(int64(2), "Grace"))
	}

	query := runtime.Query{
		Kind:       runtime.NewString("sql"),
		Expression: runtime.NewString(sqlText),
	}

	one, err := conn.QueryOne(ctx, query)
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

	count, err := conn.QueryCount(ctx, query)
	if err != nil {
		t.Fatalf("unexpected query count error: %v", err)
	}
	if count != runtime.NewInt(2) {
		t.Fatalf("expected count 2, got %v", count)
	}

	exists, err := conn.QueryExists(ctx, query)
	if err != nil {
		t.Fatalf("unexpected query exists error: %v", err)
	}
	if exists != runtime.True {
		t.Fatalf("expected exists true, got %v", exists)
	}
}

func TestConnectionQueryExecMetadataAndModifiers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, mock := mockDB(t)
	conn := NewConnection(db)

	mock.ExpectExec("UPDATE users SET name = $1 WHERE id = $2").
		WithArgs("Ada", int64(1)).
		WillReturnResult(sqlmock.NewResult(99, 3))

	out, err := conn.Query(ctx, runtime.Query{
		Kind:       runtime.NewString("SQL_EXEC"),
		Expression: runtime.NewString("UPDATE users SET name = $1 WHERE id = $2"),
		Params: runtime.NewObjectWith(map[string]runtime.Value{
			"params": runtime.NewArrayWith(runtime.NewString("Ada"), runtime.NewInt(1)),
		}),
	})
	if err != nil {
		t.Fatalf("unexpected query exec error: %v", err)
	}
	rows, ok := out.(*runtime.Array)
	if !ok {
		t.Fatalf("expected runtime array, got %T", out)
	}
	assertArrayLen(t, ctx, rows, 1)
	assertExecMetadata(t, ctx, mustObjectAt(t, ctx, rows, 0), runtime.NewInt64(3), runtime.None)

	mock.ExpectExec("DELETE FROM users WHERE id = $1").
		WithArgs(int64(99)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	count, err := conn.QueryCount(ctx, runtime.Query{
		Kind:       runtime.NewString("sql_exec"),
		Expression: runtime.NewString("DELETE FROM users WHERE id = $1"),
		Params: runtime.NewObjectWith(map[string]runtime.Value{
			"params": runtime.NewArrayWith(runtime.NewInt(99)),
		}),
	})
	if err != nil {
		t.Fatalf("unexpected query count exec error: %v", err)
	}
	if count != runtime.NewInt(1) {
		t.Fatalf("expected query count exec to return 1, got %v", count)
	}

	mock.ExpectExec("DELETE FROM users WHERE id = $1").
		WithArgs(int64(100)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	exists, err := conn.QueryExists(ctx, runtime.Query{
		Kind:       runtime.NewString("sql_exec"),
		Expression: runtime.NewString("DELETE FROM users WHERE id = $1"),
		Params: runtime.NewObjectWith(map[string]runtime.Value{
			"params": runtime.NewArrayWith(runtime.NewInt(100)),
		}),
	})
	if err != nil {
		t.Fatalf("unexpected query exists exec error: %v", err)
	}
	if exists != runtime.True {
		t.Fatalf("expected query exists exec to return true, got %v", exists)
	}
}

func TestQueryErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, mock := mockDB(t)
	conn := NewConnection(db)

	_, err := conn.Query(ctx, runtime.Query{Kind: runtime.NewString("css"), Expression: runtime.NewString("body")})
	assertErrorContains(t, err, `unsupported dialect "css"; expected "sql" or "sql_exec"`)

	mock.ExpectClose()
	if err := conn.Close(); err != nil {
		t.Fatalf("unexpected close error: %v", err)
	}

	_, err = conn.Query(ctx, runtime.Query{Kind: runtime.NewString("sql"), Expression: runtime.NewString("SELECT 1")})
	assertErrorContains(t, err, "database connection has been closed")
}
