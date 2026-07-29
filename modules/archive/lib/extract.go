package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/archive/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// Extract writes eligible archive entries through the configured filesystem.
func Extract(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 2, 3); err != nil {
		return runtime.None, err
	}

	source, err := sdk.DecodeArg[string](ctx, args, 0, sdk.RequireType(runtime.TypeString))
	if err != nil {
		return runtime.None, err
	}

	destination, err := sdk.DecodeArg[string](ctx, args, 1, sdk.RequireType(runtime.TypeString))
	if err != nil {
		return runtime.None, err
	}

	config, err := core.ConfigFrom(ctx)
	if err != nil {
		return runtime.None, err
	}

	opts, err := sdk.DecodeArgOr(
		ctx,
		args,
		2,
		core.DefaultExtractOptions(config),
		sdk.DisallowUnknownFields(),
	)

	if err != nil {
		return runtime.None, err
	}

	results, err := core.Extract(ctx, source, destination, opts)
	if err != nil {
		return runtime.None, err
	}

	return sdk.Encode(ctx, results)
}
