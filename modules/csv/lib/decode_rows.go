package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/csv/types"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// DecodeRows decodes a CSV string into raw row arrays.
// @param {String} data - CSV string.
// @param {Options} [opts] - Options for decoding.
// @return {Any[][]} - Array of row arrays.
func DecodeRows(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 2); err != nil {
		return nil, err
	}

	data, err := runtime.CastArgAt[runtime.String](args, 0)
	if err != nil {
		return nil, err
	}

	opts := types.DefaultOptions()

	if len(args) > 1 {
		optsmap, err := runtime.CastArgAt[runtime.Map](args, 1)
		if err != nil {
			return nil, err
		}

		if err := sdk.Decode(optsmap, &opts); err != nil {
			return nil, err
		}
	}

	return types.DecodeRows(ctx, data, opts)
}
