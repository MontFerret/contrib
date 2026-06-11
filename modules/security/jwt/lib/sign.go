package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/security/jwt/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func signWithConfig(cfg core.Config) func(context.Context, ...runtime.Value) (runtime.Value, error) {
	return func(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
		if err := runtime.ValidateArgs(args, 3, 3); err != nil {
			return nil, err
		}

		claimsMap, err := runtime.CastArgAt[runtime.Map](args, 0)
		if err != nil {
			return nil, core.OperationError("SIGN", err)
		}

		optsMap, err := runtime.CastArgAt[runtime.Map](args, 2)
		if err != nil {
			return nil, core.OperationError("SIGN", err)
		}

		opts, err := core.DecodeSignOptions(optsMap)
		if err != nil {
			return nil, core.OperationError("SIGN", err)
		}

		result, err := core.Sign(ctx, claimsMap, args[1], opts)
		if err != nil {
			return nil, core.OperationError("SIGN", err)
		}

		return result, nil
	}
}
