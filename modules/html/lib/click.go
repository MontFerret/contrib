package html

import (
	"context"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// CLICK dispatches click event on a given element
// @param {HTMLPage | HTMLDocument | HTMLElement} node - Target html node.
// @param {String | Int} [cssSelectorOrClicks] - CSS selector or count of clicks.
// @param {Int} [clicks=1] - Count of clicks.
func Click(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 1, 3)

	if err != nil {
		return runtime.False, err
	}

	el, err := drivers.ToElement(args[0])

	if err != nil {
		return runtime.False, err
	}

	// CLICK(elOrDoc)
	if len(args) == 1 {
		return runtime.True, el.Click(ctx, 1)
	}

	if len(args) == 2 {
		err := runtime.ValidateType(args[1], runtime.TypeString, runtime.TypeInt, runtime.TypeObject)

		if err != nil {
			return runtime.False, err
		}

		switch args[1].(type) {
		case runtime.String:
		case runtime.Map:
			selector, err := drivers.ToQuerySelector(args[1])

			if err != nil {
				return runtime.None, err
			}

			exists, err := el.ExistsBySelector(ctx, selector)

			if err != nil {
				return runtime.False, err
			}

			if !exists {
				return exists, nil
			}

			return exists, el.ClickBySelector(ctx, selector, 1)
		}

		times, err := runtime.CastInt(args[1])

		if err != nil {
			return runtime.False, err
		}

		return runtime.True, el.Click(ctx, times)
	}

	err = runtime.ValidateType(args[2], runtime.TypeInt)

	if err != nil {
		return runtime.False, err
	}

	// CLICK(doc, selector)
	selector, err := drivers.ToQuerySelector(args[1])

	if err != nil {
		return runtime.None, err
	}

	exists, err := el.ExistsBySelector(ctx, selector)

	if err != nil {
		return runtime.False, err
	}

	if !exists {
		return exists, nil
	}

	times, err := runtime.CastInt(args[1])

	if err != nil {
		return runtime.False, err
	}

	return exists, el.ClickBySelector(ctx, selector, times)
}
