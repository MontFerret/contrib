package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/csv/types"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

func DecodeRowsStream(_ context.Context, args ...runtime.Value) (runtime.Value, error) {
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

	iter := types.NewDecodeRowsIterator(content, opts)

	return sdk.NewProxy(iter), nil
}
