package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/toml/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Encode serializes a Ferret runtime object into TOML text.
// @param {Object} value - Ferret runtime value.
// @param {Object} [options] - Encode options.
// @return {String} - TOML text.
func Encode(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 2); err != nil {
		return nil, err
	}

	opts := core.DefaultEncodeOptions()
	var err error

	if len(args) > 1 {
		opts, err = core.ParseEncodeOptions(ctx, args[1])
		if err != nil {
			return nil, err
		}
	}

	text, err := core.Encode(ctx, args[0], opts)
	if err != nil {
		return nil, err
	}

	return runtime.NewString(text), nil
}
