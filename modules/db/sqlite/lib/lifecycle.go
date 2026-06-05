package lib

import (
	"context"
	"fmt"

	"github.com/MontFerret/contrib/modules/db/sqlite/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Open creates a SQLite database handle from an options object.
func Open(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	return open(ctx, core.DefaultOpenPolicy(), args...)
}

func openWithPolicy(policy core.OpenPolicy) runtime.Function {
	return func(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
		return open(ctx, policy, args...)
	}
}

func open(ctx context.Context, policy core.OpenPolicy, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 1); err != nil {
		return runtime.None, err
	}

	options, err := core.DecodeOpenOptions(args[0])
	if err != nil {
		return runtime.None, err
	}

	return core.OpenWithPolicy(ctx, options, policy)
}

// Close closes a SQLite database handle. Closing an already closed database is
// idempotent.
func Close(_ context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 1); err != nil {
		return runtime.None, err
	}

	db, ok := args[0].(*core.Connection)
	if !ok {
		return runtime.None, fmt.Errorf("DB::SQLITE CLOSE failed: expected SQLite database handle")
	}

	if err := db.Close(); err != nil {
		return runtime.None, err
	}

	return runtime.True, nil
}

// Begin starts a SQLite transaction and returns a transaction handle.
func Begin(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 1); err != nil {
		return runtime.None, err
	}

	db, ok := args[0].(*core.Connection)
	if !ok {
		return runtime.None, fmt.Errorf("DB::SQLITE BEGIN failed: expected SQLite database handle")
	}

	return db.Begin(ctx)
}

// Commit commits a SQLite transaction handle.
func Commit(_ context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 1); err != nil {
		return runtime.None, err
	}

	tx, ok := args[0].(*core.Transaction)
	if !ok {
		return runtime.None, fmt.Errorf("DB::SQLITE COMMIT failed: expected SQLite transaction handle")
	}

	if err := tx.Commit(); err != nil {
		return runtime.None, err
	}

	return runtime.True, nil
}

// Rollback rolls back a SQLite transaction handle.
func Rollback(_ context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 1); err != nil {
		return runtime.None, err
	}

	tx, ok := args[0].(*core.Transaction)
	if !ok {
		return runtime.None, fmt.Errorf("DB::SQLITE ROLLBACK failed: expected SQLite transaction handle")
	}

	if err := tx.Rollback(); err != nil {
		return runtime.None, err
	}

	return runtime.True, nil
}
