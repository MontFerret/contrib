package core

import (
	"context"
	"database/sql"

	// Register the pure-Go SQLite driver for database/sql.
	_ "modernc.org/sqlite"
)

const driverName = "sqlite"

// Open opens a SQLite database connection from validated options.
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
