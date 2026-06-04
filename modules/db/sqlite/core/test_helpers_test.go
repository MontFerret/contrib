package core

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func boolPtr(value bool) *bool {
	return &value
}

func stringPtr(value string) *string {
	return &value
}

func openMemoryDB(t *testing.T, ctx context.Context) *Connection {
	t.Helper()

	db, err := Open(ctx, OpenOptions{Memory: boolPtr(true)})
	if err != nil {
		t.Fatalf("unexpected open error: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("unexpected close error: %v", err)
		}
	})

	return db
}

func dispatchSQLForTest(t *testing.T, ctx context.Context, target interface {
	Dispatch(context.Context, runtime.DispatchEvent) (runtime.Value, error)
}, sqlText string, params ...runtime.Value) runtime.Value {
	t.Helper()

	var options runtime.Value = runtime.None
	if len(params) > 0 {
		options = runtime.NewObjectWith(map[string]runtime.Value{
			"params": runtime.NewArrayWith(params...),
		})
	}

	out, err := target.Dispatch(ctx, runtime.DispatchEvent{
		Name:    runtime.NewString("sql"),
		Payload: runtime.NewString(sqlText),
		Options: options,
	})
	if err != nil {
		t.Fatalf("unexpected dispatch error: %v", err)
	}

	return out
}

func queryRowsForTest(t *testing.T, ctx context.Context, target runtime.Queryable, sqlText string, params ...runtime.Value) *runtime.Array {
	t.Helper()

	var options runtime.Value = runtime.None
	if len(params) > 0 {
		options = runtime.NewObjectWith(map[string]runtime.Value{
			"params": runtime.NewArrayWith(params...),
		})
	}

	out, err := target.Query(ctx, runtime.Query{
		Kind:    runtime.NewString("sql"),
		Payload: runtime.NewString(sqlText),
		Options: options,
	})
	if err != nil {
		t.Fatalf("unexpected query error: %v", err)
	}

	arr, ok := out.(*runtime.Array)
	if !ok {
		t.Fatalf("expected runtime array, got %T", out)
	}

	return arr
}

func queryExecForTest(t *testing.T, ctx context.Context, target runtime.Queryable, sqlText string, params ...runtime.Value) *runtime.Object {
	t.Helper()

	var options runtime.Value = runtime.None
	if len(params) > 0 {
		options = runtime.NewObjectWith(map[string]runtime.Value{
			"params": runtime.NewArrayWith(params...),
		})
	}

	out, err := target.QueryOne(ctx, runtime.Query{
		Kind:    runtime.NewString("sql_exec"),
		Payload: runtime.NewString(sqlText),
		Options: options,
	})
	if err != nil {
		t.Fatalf("unexpected query exec error: %v", err)
	}

	obj, ok := out.(*runtime.Object)
	if !ok {
		t.Fatalf("expected object, got %T", out)
	}

	return obj
}

func assertExecMetadata(t *testing.T, ctx context.Context, obj *runtime.Object, rowsAffected, lastInsertID runtime.Value) {
	t.Helper()

	if got := objectField(t, ctx, obj, "rowsAffected"); got != rowsAffected {
		t.Fatalf("expected rowsAffected %v, got %v", rowsAffected, got)
	}
	if got := objectField(t, ctx, obj, "lastInsertId"); got != lastInsertID {
		t.Fatalf("expected lastInsertId %v, got %v", lastInsertID, got)
	}
}

func mustObjectAt(t *testing.T, ctx context.Context, arr *runtime.Array, idx int) *runtime.Object {
	t.Helper()

	value, err := arr.At(ctx, runtime.Int(idx))
	if err != nil {
		t.Fatalf("unexpected array read error: %v", err)
	}

	obj, ok := value.(*runtime.Object)
	if !ok {
		t.Fatalf("expected object, got %T", value)
	}

	return obj
}

func objectField(t *testing.T, ctx context.Context, obj *runtime.Object, name string) runtime.Value {
	t.Helper()

	value, err := obj.Get(ctx, runtime.NewString(name))
	if err != nil {
		t.Fatalf("unexpected object read error: %v", err)
	}

	return value
}

func assertArrayLen(t *testing.T, ctx context.Context, arr *runtime.Array, expected int) {
	t.Helper()

	length, err := arr.Length(ctx)
	if err != nil {
		t.Fatalf("unexpected length error: %v", err)
	}

	if int(length) != expected {
		t.Fatalf("expected array length %d, got %d", expected, int(length))
	}
}

func tempDBPath(t *testing.T) string {
	t.Helper()

	return filepath.Join(t.TempDir(), "ferret.db")
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
