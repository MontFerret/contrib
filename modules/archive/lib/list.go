package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/archive/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// List returns metadata for every entry in an archive.
//
// @param source {String} Sandboxed archive path.
// @param options {Object?} Archive format options.
// @return {Array<Object>} Archive entry metadata.
func List(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 2); err != nil {
		return runtime.None, err
	}

	source, err := sdk.DecodeArg[string](ctx, args, 0, sdk.RequireType(runtime.TypeString))
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
		1,
		core.DefaultListOptions(config),
		sdk.DisallowUnknownFields(),
	)
	if err != nil {
		return runtime.None, err
	}

	opts.Config = config
	entries, err := core.List(ctx, source, opts)
	if err != nil {
		return runtime.None, err
	}

	return sdk.Encode(ctx, entries)
}
