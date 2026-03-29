package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/csv/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// DecodeRows decodes CSV text into an array of raw row arrays.
// It keeps header rows as data and applies decoding options to each field.
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

	return core.DecodeRows(ctx, data, opts)
}
