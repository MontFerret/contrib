package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/security/jwt/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func verifyWithConfig(cfg core.Config) func(context.Context, ...runtime.Value) (runtime.Value, error) {
	return func(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
		if err := runtime.ValidateArgs(args, 3, 3); err != nil {
			return nil, err
		}

		token, err := core.ResolveToken(args[0])
		if err != nil {
			return nil, core.OperationError("VERIFY", err)
		}

		optsMap, err := runtime.CastArgAt[runtime.Map](args, 2)
		if err != nil {
			return nil, core.OperationError("VERIFY", err)
		}

		opts, err := core.DecodeVerifyOptions(optsMap)
		if err != nil {
			return nil, core.OperationError("VERIFY", err)
		}

		result, err := core.Verify(ctx, cfg, token, args[1], opts)
		if err != nil {
			return nil, core.OperationError("VERIFY", err)
		}

		return result, nil
	}
}
