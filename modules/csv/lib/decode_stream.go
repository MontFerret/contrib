package lib

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// DecodeStream decodes CSV content from string or binary input.
// It returns a proxy over an iterator of objects keyed by the original CSV
// record number after parsing.
// @param {String|Binary} data - CSV content.
// @param {Options} [opts] - Options for decoding.
// @return {Iterator<Object>} - Proxy exposing an iterator over decoded objects.
func DecodeStream(_ context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 2); err != nil {
		return nil, err
	}

	content, err := core.ResolveContent(args[0])
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

	iter, err := core.NewDecodeIterator(content, opts)
	if err != nil {
		return nil, err
	}

	return sdk.NewProxy(iter), nil
}
