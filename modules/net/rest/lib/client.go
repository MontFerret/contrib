package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/net/rest/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Client creates a reusable HTTP API client from a configuration object.
func Client(ctx context.Context, arg runtime.Value) (runtime.Value, error) {
	config, err := core.DecodeClientConfig(ctx, arg)
	if err != nil {
		return runtime.None, err
	}

	return core.NewClient(config), nil
}
