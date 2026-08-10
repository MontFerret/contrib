package lib

import (
	"context"
	"strings"

	"github.com/MontFerret/contrib/modules/archive/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// Read reads the first matching regular file from an archive.
//
// @param source {String} Sandboxed archive path.
// @param name {String} Archive entry name.
// @param options {Object?} Format, output, and missing-entry options.
// @return {String|Binary|None} Entry content, or None when configured for a missing entry.
func Read(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 2, 3); err != nil {
		return runtime.None, err
	}

	source, err := sdk.DecodeArg[string](ctx, args, 0, sdk.RequireType(runtime.TypeString))
	if err != nil {
		return runtime.None, err
	}

	name, err := sdk.DecodeArg[string](ctx, args, 1, sdk.RequireType(runtime.TypeString))
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
		core.DefaultReadOptions(config),
		sdk.DisallowUnknownFields(),
	)
	if err != nil {
		return runtime.None, err
	}

	data, found, err := core.Read(ctx, source, name, opts)
	if err != nil {
		return runtime.None, err
	}

	if !found {
		if strings.EqualFold(opts.Missing, "none") {
			return runtime.None, nil
		}

		return runtime.None, runtime.Errorf(runtime.ErrNotFound, "archive entry %q does not exist", name)
	}

	if strings.EqualFold(opts.As, "string") {
		return runtime.NewString(string(data)), nil
	}

	return runtime.NewBinary(data), nil
}
