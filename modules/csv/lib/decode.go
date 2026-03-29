package lib

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// Decode decodes CSV text into an array of objects.
// When opts.header is true, the first record defines object keys; otherwise
// opts.columns or generated colN names are used.
// @param {String} data - CSV string.
// @param {Options} [opts] - Options for decoding.
// @return {Any[]} - Array of decoded objects.
func Decode(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
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

	return core.Decode(ctx, data, opts)
}
