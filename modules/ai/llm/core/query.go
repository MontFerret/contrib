package core

import (
	"context"
	"strings"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// ExecuteQuery decodes Queryable fields and executes the scalar operation.
func ExecuteQuery(ctx context.Context, target Target, query runtime.Query) (runtime.Value, error) {
	kind := strings.ToLower(strings.TrimSpace(query.Kind.String()))
	if kind == "" {
		kind = string(ModeGenerate)
	}

	mode := Mode(kind)
	if _, err := semanticKeys(mode); err != nil {
		return runtime.None, OperationError("QUERY", err)
	}

	semantic, err := DecodeSemanticOptions(ctx, mode, query.Params)
	if err != nil {
		return runtime.None, OperationError("QUERY", err)
	}

	execution, err := DecodeExecutionOptions(ctx, query.Options)
	if err != nil {
		return runtime.None, OperationError("QUERY", err)
	}

	value, err := Execute(ctx, target, OperationRequest{
		Mode:      mode,
		Input:     query.Expression.String(),
		Semantic:  semantic,
		Execution: execution,
	})

	if err != nil {
		return runtime.None, OperationError("QUERY", err)
	}

	return value, nil
}
