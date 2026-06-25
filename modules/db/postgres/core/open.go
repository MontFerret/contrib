package core

import (
	"context"
	"database/sql"

	// Register pgx's database/sql driver.
	_ "github.com/jackc/pgx/v5/stdlib"
)

const driverName = "pgx"

// Open opens a Postgres database connection from validated options.
func Open(ctx context.Context, options OpenOptions) (*Connection, error) {
	dsn, err := options.dsn()
	if err != nil {
		return nil, OperationError("OPEN", err)
	}

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, OperationError("OPEN", err)
	}

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()

		return nil, OperationError("OPEN", err)
	}

	return NewConnection(db), nil
}
