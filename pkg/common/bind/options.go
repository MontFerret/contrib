package bind

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// DecodeMapArgOrDefault decodes an optional map argument into a copy of defaults.
func DecodeMapArgOrDefault[T any](ctx context.Context, args []runtime.Value, index int, defaults T) (T, error) {
	if index >= len(args) {
		return defaults, nil
	}

	arg, err := runtime.CastArgAt[runtime.Map](args, index)
	if err != nil {
		return defaults, err
	}

	out := defaults
	if err := sdk.Decode(ctx, arg, &out); err != nil {
		return defaults, err
	}

	return out, nil
}
