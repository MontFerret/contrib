package core

import (
	"context"
	"database/sql"
)

type (
	queryRunner interface {
		QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	}

	execRunner interface {
		ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	}

	sqlRunner interface {
		queryRunner
		execRunner
	}
)
