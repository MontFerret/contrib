package bind

import (
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// DecodeMapArgOrDefault decodes an optional map argument into a copy of defaults.
func DecodeMapArgOrDefault[T any](args []runtime.Value, index int, defaults T) (T, error) {
	if index >= len(args) {
		return defaults, nil
	}

	arg, err := runtime.CastArgAt[runtime.Map](args, index)
	if err != nil {
		return defaults, err
	}

	out := defaults
	if err := sdk.Decode(arg, &out); err != nil {
		return defaults, err
	}

	return out, nil
}
