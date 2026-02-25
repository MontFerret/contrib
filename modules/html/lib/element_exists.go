package html

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// ELEMENT_EXISTS returns a boolean value indicating whether there is an element matched by selector.
// @param {HTMLPage | HTMLDocument | HTMLElement} node - Target html node.
// @param {String} selector - CSS selector.
// @return {Boolean} - A boolean value indicating whether there is an element matched by selector.
func ElementExists(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	el, selector, err := queryArgs(args)

	if err != nil {
		return runtime.None, err
	}

	return el.ExistsBySelector(ctx, selector)
}
