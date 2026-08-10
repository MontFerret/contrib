package lib

import (
	"context"
	"fmt"

	"github.com/MontFerret/contrib/modules/db/sqlite/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Open creates a SQLite database handle from connection options.
func Open(ctx context.Context, arg runtime.Value) (runtime.Value, error) {
	return open(ctx, core.DefaultOpenPolicy(), arg)
}

// openWithPolicy creates a SQLite database handle under the configured policy.
//
// @param options {Object} SQLite connection options.
// @return {SQLiteDatabase} Open database handle.
func openWithPolicy(policy core.OpenPolicy) runtime.Function1 {
	return func(ctx context.Context, arg runtime.Value) (runtime.Value, error) {
		return open(ctx, policy, arg)
	}
}

func open(ctx context.Context, policy core.OpenPolicy, arg runtime.Value) (runtime.Value, error) {
	options, err := core.DecodeOpenOptions(ctx, arg)
	if err != nil {
		return runtime.None, err
	}

	return core.OpenWithPolicy(ctx, options, policy)
}

// Close closes a SQLite database handle. Closing an already closed database is
// idempotent.
//
// @param database {SQLiteDatabase} Database handle to close.
// @return {Boolean} True when the handle is closed.
func Close(_ context.Context, arg runtime.Value) (runtime.Value, error) {
	db, ok := arg.(*core.Connection)
	if !ok {
		return runtime.None, fmt.Errorf("DB::SQLITE CLOSE failed: expected SQLite database handle")
	}

	if err := db.Close(); err != nil {
		return runtime.None, err
	}

	return runtime.True, nil
}

// Begin starts a SQLite transaction and returns a transaction handle.
//
// @param database {SQLiteDatabase} Database handle.
// @return {SQLiteTransaction} Open transaction handle.
func Begin(ctx context.Context, arg runtime.Value) (runtime.Value, error) {
	db, ok := arg.(*core.Connection)
	if !ok {
		return runtime.None, fmt.Errorf("DB::SQLITE BEGIN failed: expected SQLite database handle")
	}

	return db.Begin(ctx)
}

// Commit commits a SQLite transaction handle.
//
// @param transaction {SQLiteTransaction} Transaction to commit.
// @return {Boolean} True when the transaction is committed.
func Commit(_ context.Context, arg runtime.Value) (runtime.Value, error) {
	tx, ok := arg.(*core.Transaction)
	if !ok {
		return runtime.None, fmt.Errorf("DB::SQLITE COMMIT failed: expected SQLite transaction handle")
	}

	if err := tx.Commit(); err != nil {
		return runtime.None, err
	}

	return runtime.True, nil
}

// Rollback rolls back a SQLite transaction handle.
//
// @param transaction {SQLiteTransaction} Transaction to roll back.
// @return {Boolean} True when the transaction is rolled back.
func Rollback(_ context.Context, arg runtime.Value) (runtime.Value, error) {
	tx, ok := arg.(*core.Transaction)
	if !ok {
		return runtime.None, fmt.Errorf("DB::SQLITE ROLLBACK failed: expected SQLite transaction handle")
	}

	if err := tx.Rollback(); err != nil {
		return runtime.None, err
	}

	return runtime.True, nil
}
