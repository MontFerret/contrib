package core

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func dispatchSQL(ctx context.Context, operation string, runner execRunner, event runtime.DispatchEvent) (runtime.Value, error) {
	if err := validateDialect(event.Name.String()); err != nil {
		return runtime.None, OperationError(operation, err)
	}

	sqlText, err := runtime.CastString(event.Payload)
	if err != nil {
		return runtime.None, OperationError(operation, err)
	}

	params, err := parseParams(ctx, event.Options)
	if err != nil {
		return runtime.None, OperationError(operation, err)
	}

	result, err := runner.ExecContext(ctx, sqlText.String(), params...)
	if err != nil {
		return runtime.None, OperationError(operation, err)
	}

	return dispatchResult(sqlText.String(), result), nil
}
