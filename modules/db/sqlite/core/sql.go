package core

import (
	"context"
	"database/sql"
)

type queryRunner interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

type execRunner interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type sqlRunner interface {
	queryRunner
	execRunner
}
