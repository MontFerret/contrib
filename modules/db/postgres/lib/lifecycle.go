package lib

import (
	"context"
	"fmt"

	"github.com/MontFerret/contrib/modules/db/postgres/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Open creates a Postgres database handle from an options object.
func Open(ctx context.Context, arg runtime.Value) (runtime.Value, error) {
	options, err := core.DecodeOpenOptions(arg)
	if err != nil {
		return runtime.None, err
	}

	return core.Open(ctx, options)
}

// Close closes a Postgres database handle. Closing an already closed database
// is idempotent.
func Close(_ context.Context, arg runtime.Value) (runtime.Value, error) {
	db, ok := arg.(*core.Connection)
	if !ok {
		return runtime.None, fmt.Errorf("DB::POSTGRES CLOSE failed: expected Postgres database handle")
	}

	if err := db.Close(); err != nil {
		return runtime.None, err
	}

	return runtime.True, nil
}

// Begin starts a Postgres transaction and returns a transaction handle.
func Begin(ctx context.Context, arg runtime.Value) (runtime.Value, error) {
	db, ok := arg.(*core.Connection)
	if !ok {
		return runtime.None, fmt.Errorf("DB::POSTGRES BEGIN failed: expected Postgres database handle")
	}

	return db.Begin(ctx)
}

// Commit commits a Postgres transaction handle.
func Commit(_ context.Context, arg runtime.Value) (runtime.Value, error) {
	tx, ok := arg.(*core.Transaction)
	if !ok {
		return runtime.None, fmt.Errorf("DB::POSTGRES COMMIT failed: expected Postgres transaction handle")
	}

	if err := tx.Commit(); err != nil {
		return runtime.None, err
	}

	return runtime.True, nil
}

// Rollback rolls back a Postgres transaction handle.
func Rollback(_ context.Context, arg runtime.Value) (runtime.Value, error) {
	tx, ok := arg.(*core.Transaction)
	if !ok {
		return runtime.None, fmt.Errorf("DB::POSTGRES ROLLBACK failed: expected Postgres transaction handle")
	}

	if err := tx.Rollback(); err != nil {
		return runtime.None, err
	}

	return runtime.True, nil
}
