package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/csv/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// Encode encodes an array of objects or row arrays into CSV text.
// Object input can emit a header row depending on opts.header and opts.columns.
// @param {Any[]} data - Array of objects or arrays.
// @param {Options} [opts] - Options for encoding.
// @return {String} - CSV text.
func Encode(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 2); err != nil {
		return nil, err
	}

	opts := core.DefaultOptions()

	if len(args) > 1 {
		optsmap, err := runtime.CastArgAt[runtime.Map](args, 1)
		if err != nil {
			return nil, err
		}

		if err := sdk.Decode(optsmap, &opts); err != nil {
			return nil, err
		}
	}

	result, err := core.Encode(ctx, args[0], opts)
	if err != nil {
		return nil, err
	}

	return runtime.NewString(result.Text), nil
}
