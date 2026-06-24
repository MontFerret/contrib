package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/csv/core"
	"github.com/MontFerret/contrib/pkg/common/bind"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
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

	opts, err := bind.DecodeMapArgOrDefault(args, 1, core.DefaultOptions())
	if err != nil {
		return nil, err
	}

	return core.DecodeRows(ctx, data, opts)
}
