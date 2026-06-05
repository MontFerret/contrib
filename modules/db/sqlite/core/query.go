package core

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func querySQL(ctx context.Context, operation string, runner sqlRunner, q runtime.Query) (runtime.List, error) {
	dialect, err := parseQueryDialect(string(q.Kind))
	if err != nil {
		return nil, OperationError(operation, err)
	}

	switch dialect {
	case queryDialectRows:
		return queryRows(ctx, operation, runner, q)
	case queryDialectExec:
		return queryExec(ctx, operation, runner, q)
	default:
		return nil, OperationErrorf(operation, "unsupported dialect %q", q.Kind.String())
	}
}

func queryRows(ctx context.Context, operation string, runner queryRunner, q runtime.Query) (runtime.List, error) {
	if err := validateDialect(string(q.Kind)); err != nil {
		return nil, OperationError(operation, err)
	}

	params, err := parseParams(ctx, q.Options)
	if err != nil {
		return nil, OperationError(operation, err)
	}

	rows, err := runner.QueryContext(ctx, q.Payload.String(), params...)
	if err != nil {
		return nil, OperationError(operation, err)
	}
	defer rows.Close()

	out, err := scanRows(ctx, rows)
	if err != nil {
		return nil, OperationError(operation, err)
	}

	return out, nil
}

func queryExec(ctx context.Context, operation string, runner execRunner, q runtime.Query) (runtime.List, error) {
	params, err := parseParams(ctx, q.Options)
	if err != nil {
		return nil, OperationError(operation, err)
	}

	result, err := runner.ExecContext(ctx, q.Payload.String(), params...)
	if err != nil {
		return nil, OperationError(operation, err)
	}

	return runtime.NewArrayWith(execResult(q.Payload.String(), result)), nil
}
