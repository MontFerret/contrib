package html

import (
	"context"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// INPUT_CLEAR clears a value from an underlying input element.
// @param {HTMLPage | HTMLDocument | HTMLElement} node - Target html node.
// @param {String} [selector] - CSS selector.
func InputClear(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 1, 2)

	if err != nil {
		return runtime.None, err
	}

	el, err := drivers.ToElement(args[0])

	if err != nil {
		return runtime.None, err
	}

	// CLEAR(el)
	if len(args) == 1 {
		return runtime.None, el.Clear(ctx)
	}

	selector, err := drivers.ToQuerySelector(args[1])

	if err != nil {
		return runtime.None, err
	}

	return runtime.True, el.ClearBySelector(ctx, selector)
}
