package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/csv/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// DecodeStream decodes CSV content from string or binary input.
// It returns an iterator value over objects keyed by the original CSV
// record number after parsing.
// @param {String|Binary} data - CSV content.
// @param {Options} [opts] - Options for decoding.
// @return {Iterator<Object>} - Iterator over decoded objects.
func DecodeStream(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 2); err != nil {
		return nil, err
	}

	content, err := core.ResolveContent(args[0])
	if err != nil {
		return nil, err
	}

	opts, err := sdk.DecodeArgOr(
		ctx,
		args,
		1,
		core.DefaultOptions(),
		sdk.DisallowUnknownFields(),
	)
	if err != nil {
		return nil, err
	}

	iter, err := core.NewDecodeIterator(content, opts)
	if err != nil {
		return nil, err
	}

	return sdk.NewIteratorValue(iter), nil
}
