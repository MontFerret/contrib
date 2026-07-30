package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// GetInnerTextAll returns an array of inner text values for matched elements.
// @param {HTMLPage | HTMLDocument | HTMLElement} node - Target html node.
// @param {String} selector - String of CSS selector.
// @return {String[]} - An array of inner text if all matched elements, otherwise empty array.
func GetInnerTextAll(ctx context.Context, root, selectorValue runtime.Value) (runtime.Value, error) {
	target, err := toRootContentTarget(root)

	if err != nil {
		return runtime.None, err
	}

	selector, err := drivers.ToQuerySelector(ctx, selectorValue)

	if err != nil {
		return runtime.None, err
	}

	return target.GetInnerTextBySelectorAll(ctx, selector)
}
