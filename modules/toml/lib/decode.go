package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/toml/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Decode decodes a TOML document into a Ferret runtime object.
// @param {String|Binary} data - TOML content.
// @param {Object} [options] - Decode options.
// @return {Object} - Decoded TOML object.
func Decode(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 2); err != nil {
		return nil, err
	}

	content, err := core.ResolveContent(args[0])
	if err != nil {
		return nil, err
	}

	opts := core.DefaultDecodeOptions()
	if len(args) > 1 {
		opts, err = core.ParseDecodeOptions(ctx, args[1])
		if err != nil {
			return nil, err
		}
	}

	return core.Decode(ctx, content, opts)
}
