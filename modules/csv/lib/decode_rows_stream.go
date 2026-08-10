package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/csv/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// DecodeRowsStream decodes CSV content from string or binary input.
// It returns an iterator value over row arrays keyed by the original CSV
// record number after parsing.
//
// @param data {String|Binary} CSV content.
// @param opts {Object?} Decoding options.
// @return {Iterator<Array<Any>>} Iterator over decoded row arrays.
func DecodeRowsStream(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
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

	iter, err := core.NewDecodeRowsIterator(content, opts)
	if err != nil {
		return nil, err
	}

	return sdk.NewIteratorValue(iter), nil
}
