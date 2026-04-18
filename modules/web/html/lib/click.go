package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Click dispatches a click event on a given element.
// @param {HTMLPage | HTMLDocument | HTMLElement} node - Target html node.
// @param {String | Int} [cssSelectorOrClicks] - CSS selector or count of clicks.
// @param {Int} [clicks=1] - Count of clicks.
func Click(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 3); err != nil {
		return runtime.False, err
	}

	target, err := drivers.ToInteractionTarget(args[0])

	if err != nil {
		return runtime.False, err
	}

	queryTarget, err := drivers.ToQueryTarget(args[0])
	if err != nil {
		return runtime.False, err
	}

	// CLICK(elOrDoc)
	if len(args) == 1 {
		return runtime.True, target.Click(ctx, 1)
	}

	if len(args) == 2 {
		err := runtime.ValidateArgType(args[1], 1, runtime.TypeString, runtime.TypeInt, runtime.TypeObject, drivers.TypeQuerySelector)

		if err != nil {
			return runtime.False, err
		}

		switch arg2 := args[1].(type) {
		case runtime.Int:
			return runtime.True, target.Click(ctx, arg2)
		default:
			selector, err := drivers.ToQuerySelector(args[1])

			if err != nil {
				return runtime.None, err
			}

			exists, err := queryTarget.ExistsBySelector(ctx, selector)

			if err != nil {
				return runtime.False, err
			}

			if !exists {
				return exists, nil
			}

			return exists, target.ClickBySelector(ctx, selector, 1)
		}
	}

	err = runtime.ValidateType(args[2], runtime.TypeInt)

	if err != nil {
		return runtime.False, err
	}

	// CLICK(doc, selector, clicks)
	selector, err := drivers.ToQuerySelector(args[1])

	if err != nil {
		return runtime.None, err
	}

	exists, err := queryTarget.ExistsBySelector(ctx, selector)

	if err != nil {
		return runtime.False, err
	}

	if !exists {
		return exists, nil
	}

	times, err := runtime.CastInt(args[2])

	if err != nil {
		return runtime.False, err
	}

	return exists, target.ClickBySelector(ctx, selector, times)
}
