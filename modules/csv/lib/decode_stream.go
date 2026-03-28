package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/csv/types"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// DecodeStream decodes CSV content from any value supported by types.ResolveContent,
// including strings, binary values, and stringer-backed host values.
// @param {String|Binary|Any} data - CSV content or value resolvable to CSV text.
// @param {Options} [opts] - Options for decoding.
// @return {Iterator<Object>} - Proxy exposing an iterator over decoded objects.
func DecodeStream(_ context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 2); err != nil {
		return nil, err
	}

	content, err := types.ResolveContent(args[0])
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

	iter, err := types.NewDecodeIterator(content, opts)
	if err != nil {
		return nil, err
	}

	return sdk.NewProxy(iter), nil
}
