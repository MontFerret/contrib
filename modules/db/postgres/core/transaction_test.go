package core

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestTransactionQueryExecMetadata(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, mock := mockDB(t)
	conn := NewConnection(db)

	mock.ExpectBegin()
	committed, err := conn.Begin(ctx)
	if err != nil {
		t.Fatalf("unexpected begin error: %v", err)
	}
	mock.ExpectExec("INSERT INTO users(name) VALUES ($1)").
		WithArgs("Ada").
		WillReturnResult(sqlmock.NewResult(0, 1))
	insert := queryExecForTest(t, ctx, committed, "INSERT INTO users(name) VALUES ($1)", runtime.NewString("Ada"))
	assertExecMetadata(t, ctx, insert, runtime.NewInt64(1), runtime.None)
	mock.ExpectCommit()
	if err := committed.Commit(); err != nil {
		t.Fatalf("unexpected commit error: %v", err)
	}

	mock.ExpectBegin()
	rolledBack, err := conn.Begin(ctx)
	if err != nil {
		t.Fatalf("unexpected begin error: %v", err)
	}
	mock.ExpectExec("INSERT INTO users(name) VALUES ($1)").
		WithArgs("Grace").
		WillReturnResult(sqlmock.NewResult(0, 1))
	queryExecForTest(t, ctx, rolledBack, "INSERT INTO users(name) VALUES ($1)", runtime.NewString("Grace"))
	mock.ExpectRollback()
	if err := rolledBack.Rollback(); err != nil {
		t.Fatalf("unexpected rollback error: %v", err)
	}
}

func TestConnectionCloseRollsBackTrackedTransactions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, mock := mockDB(t)
	conn := NewConnection(db)

	mock.ExpectBegin()
	if _, err := conn.Begin(ctx); err != nil {
		t.Fatalf("unexpected begin error: %v", err)
	}

	mock.ExpectRollback()
	mock.ExpectClose()
	if err := conn.Close(); err != nil {
		t.Fatalf("unexpected close error: %v", err)
	}
}

func TestTransactionErrorsAfterFinishAndParentClose(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, mock := mockDB(t)
	conn := NewConnection(db)

	mock.ExpectBegin()
	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatalf("unexpected begin error: %v", err)
	}
	mock.ExpectCommit()
	if err := tx.Commit(); err != nil {
		t.Fatalf("unexpected commit error: %v", err)
	}
	err = tx.Commit()
	assertErrorContains(t, err, "transaction has already been finished")

	mock.ExpectBegin()
	tx, err = conn.Begin(ctx)
	if err != nil {
		t.Fatalf("unexpected begin error: %v", err)
	}
	mock.ExpectRollback()
	mock.ExpectClose()
	if err := conn.Close(); err != nil {
		t.Fatalf("unexpected close error: %v", err)
	}

	_, err = tx.Query(ctx, runtime.Query{Kind: runtime.NewString("sql"), Expression: runtime.NewString("SELECT 1")})
	assertErrorContains(t, err, "parent database has been closed")
}
