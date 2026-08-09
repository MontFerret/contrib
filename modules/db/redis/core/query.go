package core

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode"

	goredis "github.com/redis/go-redis/v9"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func executeQuery(ctx context.Context, connection *Connection, q runtime.Query) (runtime.Value, bool, error) {
	dialect, err := parseQueryDialect(q.Kind.String())
	if err != nil {
		return runtime.None, false, OperationError("QUERY", err)
	}

	command, err := parseCommand(q.Expression.String())
	if err != nil {
		return runtime.None, false, OperationError("QUERY", err)
	}

	params, err := parseParams(ctx, q.Params)
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

	args := make([]any, 1, len(params)+1)
	args[0] = command
	args = append(args, params...)

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

func parseCommand(input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("redis command must not be empty")
	}

	if strings.TrimSpace(input) != input || strings.IndexFunc(input, unicode.IsSpace) >= 0 {
		return "", fmt.Errorf("redis command must be a single token; put subcommands and modifiers in WITH.params")
	}

	return input, nil
}
