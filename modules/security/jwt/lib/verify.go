package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/security/jwt/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// verifyWithConfig verifies a compact JWT signature and claims.
//
// @param token {String|Binary} Compact JWT.
// @param key {JWTKey} Verification key handle.
// @param options {Object} Verification and claim requirements.
// @return {Object} Verified header, claims, and verification state.
func verifyWithConfig(cfg core.Config) func(context.Context, ...runtime.Value) (runtime.Value, error) {
	return func(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
		if err := runtime.ValidateArgs(args, 3, 3); err != nil {
			return nil, err
		}

		token, err := core.ResolveToken(args[0])
		if err != nil {
			return nil, core.OperationError("VERIFY", err)
		}

		optsMap, err := sdk.DecodeArg[runtime.Map](ctx, args, 2)
		if err != nil {
			return nil, core.OperationError("VERIFY", err)
		}

		opts, err := core.DecodeVerifyOptions(ctx, optsMap)
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
