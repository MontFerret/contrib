package core

import (
	"context"
	"errors"

	goredis "github.com/redis/go-redis/v9"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func executeQuery(ctx context.Context, connection *Connection, q runtime.Query) (runtime.Value, bool, error) {
	dialect, err := parseQueryDialect(q.Kind.String())
	if err != nil {
		return runtime.None, false, OperationError("QUERY", err)
	}

	command, commandArgs, err := compileQueryTemplate(ctx, q.Expression.String(), q.Params)
	if err != nil {
		return runtime.None, false, OperationError("QUERY", err)
	}

	options, err := decodeExecutionOptions(ctx, q.Options)
	if err != nil {
		return runtime.None, false, OperationError("QUERY", err)
	}

	queryCtx := ctx
	cancel := func() {}

	if options.Timeout > 0 {
		queryCtx, cancel = context.WithTimeout(ctx, options.Timeout)
	}

	defer cancel()

	client, err := connection.redisClient("QUERY")
	if err != nil {
		return runtime.None, false, err
	}

	if dialect == queryDialectRead {
		info, err := connection.readCommandInfo(queryCtx, client, command)
		if err != nil {
			return runtime.None, false, OperationError("QUERY", err)
		}

		if !info.ReadOnly {
			return runtime.None, false, OperationErrorf(
				"QUERY",
				"command %q is not marked read-only; use redis_exec",
				command,
			)
		}
	}

	args := make([]any, 1, len(commandArgs)+1)
	args[0] = command
	args = append(args, commandArgs...)

	result, err := client.Do(queryCtx, args...).Result()
	if errors.Is(err, goredis.Nil) {
		return runtime.None, false, nil
	}

	if err != nil {
		return runtime.None, false, OperationError("QUERY", err)
	}

	value, err := redisValueToRuntime(queryCtx, result)
	if err != nil {
		return runtime.None, false, OperationError("QUERY", err)
	}

	_, flatten := value.(runtime.List)

	return value, flatten, nil
}
