package html

import (
	"context"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// SELECT selects a value from an underlying select element.
// @param {HTMLElement} element - Target html element.
// @param {String | String[]} valueOrSelector - Selector or a an array of strings as a value.
// @param {String[]} value - Target value. Optional.
// @return {String[]} - Array of selected runtime.
func Select(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 2, 4)

	if err != nil {
		return runtime.None, err
	}

	el, err := drivers.ToElement(args[0])

	if err != nil {
		return runtime.None, err
	}

	if len(args) == 2 {
		arr, err := runtime.ToList(ctx, args[1])

		if err != nil {
			return runtime.None, err
		}

		return el.Select(ctx, arr)
	}

	selector, err := drivers.ToQuerySelector(args[1])

	if err != nil {
		return runtime.None, err
	}

	arr, err := runtime.ToList(ctx, args[2])

	if err != nil {
		return runtime.None, err
	}

	return el.SelectBySelector(ctx, selector, arr)
}
