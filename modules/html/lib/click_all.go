package html

import (
	"context"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// CLICK_ALL dispatches click event on all matched element
// @param {HTMLPage | HTMLDocument | HTMLElement} node - Target html node.
// @param {String} selector - CSS selector.
// @param {Int} [clicks=1] - Optional count of clicks.
// @return {Boolean} - True if matched at least one element.
func ClickAll(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 2, 3)

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

	exists, err := el.ExistsBySelector(ctx, selector)

	if err != nil {
		return runtime.False, err
	}

	if !exists {
		return runtime.False, nil
	}

	times := runtime.NewInt(1)

	if len(args) == 3 {
		err := runtime.ValidateType(args[2], runtime.TypeInt)

		if err != nil {
			return runtime.False, err
		}

		times, err = runtime.CastInt(args[2])

		if err != nil {
			return runtime.False, err
		}
	}

	return runtime.True, el.ClickBySelectorAll(ctx, selector, times)
}
