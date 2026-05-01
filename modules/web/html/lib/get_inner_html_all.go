package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// GetInnerHTMLAll returns an array of inner HTML strings of matched elements.
// @param {HTMLPage | HTMLDocument | HTMLElement} node - Target html node.
// @param {String} selector - String of CSS selector.
// @return {String[]} - An array of inner HTML strings if all matched elements, otherwise empty array.
func GetInnerHTMLAll(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 2, 2)

	if err != nil {
		return runtime.None, err
	}

	target, err := toRootContentTarget(args[0])

	if err != nil {
		return runtime.None, err
	}

	selector, err := drivers.ToQuerySelector(args[1])

	if err != nil {
		return runtime.None, err
	}

	return target.GetInnerHTMLBySelectorAll(ctx, selector)
}
