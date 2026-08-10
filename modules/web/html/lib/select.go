package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Select selects values in a select element.
//
// @param root {HTMLPage|HTMLDocument|HTMLElement} HTML root.
// @param valuesOrSelector {String|Array<String>} Values or element selector.
// @param values {String|Array<String>?} Values when a selector is supplied.
// @return {Array<String>} Selected values.
func Select(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 2, 4)

	if err != nil {
		return runtime.None, err
	}

	target, err := toRootInteractionTarget(args[0])
	if err != nil {
		return runtime.None, err
	}

	if len(args) == 2 {
		arr, err := runtime.ToList(ctx, args[1])

		if err != nil {
			return runtime.None, err
		}

		return target.Select(ctx, arr)
	}

	selector, err := drivers.ToQuerySelector(ctx, args[1])

	if err != nil {
		return runtime.None, err
	}

	arr, err := runtime.ToList(ctx, args[2])

	if err != nil {
		return runtime.None, err
	}

	return target.SelectBySelector(ctx, selector, arr)
}
