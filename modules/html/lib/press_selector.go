package html

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"

	"github.com/MontFerret/contrib/modules/html/drivers"
)

// PRESS_SELECTOR presses a keyboard key.
// @param {HTMLPage | HTMLDocument | HTMLElement} node - Target html node.
// @param {String} selector - CSS selector.
// @param {String | String[]} key - Target keyboard key(s).
// @param {Int} [presses=1] - Count of presses.
func PressSelector(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 3, 4)

	if err != nil {
		return runtime.False, err
	}

	el, err := drivers.ToElement(args[0])

	if err != nil {
		return runtime.False, err
	}

	selector, err := drivers.ToQuerySelector(args[1])

	if err != nil {
		return runtime.None, err
	}

	count := runtime.NewInt(1)

	if len(args) == 4 {
		countArg, err := runtime.ToInt(ctx, args[3])

		if err != nil {
			return runtime.None, err
		}

		if countArg > 0 {
			count = countArg
		}
	}

	keysArg := args[2]

	switch keys := keysArg.(type) {
	case runtime.String:
		return runtime.True, el.PressBySelector(ctx, selector, []runtime.String{keys}, count)
	case runtime.List:
		keySlice, err := sdk.ToSlice(ctx, keys, func(ctx context.Context, value, key runtime.Value) (runtime.String, error) {
			return runtime.ToString(value), nil
		})

		if err != nil {
			return runtime.None, err
		}

		return runtime.True, el.PressBySelector(ctx, selector, keySlice, count)
	default:
		return runtime.None, runtime.TypeErrorOf(keysArg, runtime.TypeString, runtime.TypeArray)
	}
}
