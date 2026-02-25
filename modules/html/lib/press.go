package html

import (
	"context"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// PRESS presses a keyboard key.
// @param {HTMLPage | HTMLDocument | HTMLElement} node - Target html node.
// @param {String | String[]} key - Target keyboard key(s).
// @param {Int} [presses=1] - Count of presses.
func Press(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 2, 3)

	if err != nil {
		return runtime.False, err
	}

	el, err := drivers.ToElement(args[0])

	if err != nil {
		return runtime.False, err
	}

	count := runtime.NewInt(1)

	if len(args) == 3 {
		countArg, err := runtime.ToInt(ctx, args[2])

		if err != nil {
			return runtime.None, err
		}

		if countArg > 0 {
			count = countArg
		}
	}

	keysArg := args[1]

	switch keys := keysArg.(type) {
	case runtime.String:
		return runtime.True, el.Press(ctx, []runtime.String{keys}, count)
	case runtime.List:
		keySlice, err := sdk.ToSlice(ctx, keys, func(ctx context.Context, value, key runtime.Value) (runtime.String, error) {
			return runtime.ToString(value), nil
		})

		if err != nil {
			return runtime.None, err
		}

		return runtime.True, el.Press(ctx, keySlice, count)
	default:
		return runtime.None, runtime.TypeErrorOf(keysArg, runtime.TypeString, runtime.TypeArray)
	}
}
