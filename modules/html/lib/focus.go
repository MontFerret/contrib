package html

import (
	"context"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// FOCUS Sets focus on the element.
// @param {HTMLPage | HTMLDocument | HTMLElement} node - Target html node.
// @param {String} [selector] - CSS selector.
func Focus(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 1, 2)

	if err != nil {
		return runtime.None, err
	}

	el, err := drivers.ToElement(args[0])

	if err != nil {
		return runtime.None, err
	}

	if len(args) == 1 {
		return runtime.True, el.Focus(ctx)
	}

	selector, err := drivers.ToQuerySelector(args[1])

	if err != nil {
		return runtime.None, err
	}

	return runtime.True, el.FocusBySelector(ctx, selector)
}
