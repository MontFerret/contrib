package lib

import (
	"context"
	"fmt"

	"github.com/MontFerret/contrib/modules/db/redis/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Open creates a Redis connection handle from connection options.
//
// @param options {Object} Redis connection options.
// @return {RedisConnection} Open connection handle.
func Open(ctx context.Context, arg runtime.Value) (runtime.Value, error) {
	options, err := core.DecodeOpenOptions(ctx, arg)
	if err != nil {
		return runtime.None, err
	}

	return core.Open(ctx, options)
}

// Close closes a Redis connection handle. Closing an already closed connection is idempotent.
//
// @param connection {RedisConnection} Connection handle to close.
// @return {Boolean} True when the handle is closed.
func Close(_ context.Context, arg runtime.Value) (runtime.Value, error) {
	connection, ok := arg.(*core.Connection)
	if !ok {
		return runtime.None, fmt.Errorf("DB::REDIS CLOSE failed: expected Redis connection handle")
	}

	if err := connection.Close(); err != nil {
		return runtime.None, err
	}

	return runtime.True, nil
}
