package lib

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
)

// PressSelector sends keyboard input to a selected element.
//
// @param root {HTMLPage|HTMLDocument|HTMLElement} HTML root.
// @param selector {String} Element selector.
// @param keys {String|Array<String>} Keyboard key or keys.
// @param count {Int?} Number of presses.
// @return {Boolean} True when the keys are sent.
func PressSelector(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 3, 4)

	if err != nil {
		return runtime.False, err
	}

	target, err := toRootInteractionTarget(args[0])

	if err != nil {
		return runtime.False, err
	}

	selector, err := drivers.ToQuerySelector(ctx, args[1])

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
		return runtime.True, target.PressBySelector(ctx, selector, []runtime.String{keys}, count)
	case runtime.List:
		keySlice, err := sdk.ToSlice(ctx, keys, func(ctx context.Context, value, key runtime.Value) (runtime.String, error) {
			return runtime.ToString(value), nil
		})

		if err != nil {
			return runtime.None, err
		}

		return runtime.True, target.PressBySelector(ctx, selector, keySlice, count)
	default:
		return runtime.None, runtime.TypeErrorOf(keysArg, runtime.TypeString, runtime.TypeArray)
	}
}
