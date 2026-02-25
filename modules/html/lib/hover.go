package html

import (
	"context"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// HOVER fetches an element with selector, scrolls it into view if needed, and then uses page.mouse to hover over the center of the element.
// If there's no element matching selector, the method returns an error.
// @param {HTMLPage | HTMLDocument | HTMLElement} node - Target html node.
// @param {String} [selector] - If document is passed, this param must represent an element selector.
func Hover(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 1, 2)

	if err != nil {
		return runtime.None, err
	}

	el, err := drivers.ToElement(args[0])

	if err != nil {
		return runtime.None, err
	}

	if len(args) == 1 {
		return runtime.True, el.Hover(ctx)
	}

	selector, err := drivers.ToQuerySelector(args[1])

	if err != nil {
		return runtime.None, err
	}

	return runtime.True, el.HoverBySelector(ctx, selector)
}
