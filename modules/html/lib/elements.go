package html

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// ELEMENTS finds HTML elements by a given CSS selector.
// Returns an empty array if element not found.
// @param {HTMLPage | HTMLDocument | HTMLElement} node - Target html node.
// @param {String} selector - CSS selector.
// @return {HTMLElement[]} - An array of matched HTML elements.
func Elements(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	el, selector, err := queryArgs(args)

	if err != nil {
		return runtime.None, err
	}

	return el.QuerySelectorAll(ctx, selector)
}
