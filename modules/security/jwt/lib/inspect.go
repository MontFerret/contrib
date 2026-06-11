package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/security/jwt/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func inspectWithConfig(cfg core.Config) func(context.Context, ...runtime.Value) (runtime.Value, error) {
	return func(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
		if err := runtime.ValidateArgs(args, 1, 1); err != nil {
			return nil, err
		}

		token, err := core.ResolveToken(args[0])
		if err != nil {
			return nil, core.OperationError("INSPECT", err)
		}

		result, err := core.Inspect(ctx, cfg, token)
		if err != nil {
			return nil, core.OperationError("INSPECT", err)
		}

		return result, nil
	}
}
