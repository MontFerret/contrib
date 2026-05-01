package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// GetInnerText returns the inner text string of a given or matched CSS selector element.
// @param {HTMLPage | HTMLDocument | HTMLElement} node - Target html node.
// @param {String} [selector] - String of CSS selector.
// @return {String} - Inner text if a matched element, otherwise empty string.
func GetInnerText(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 1, 2)

	if err != nil {
		return runtime.EmptyString, err
	}

	target, err := toRootContentTarget(args[0])

	if err != nil {
		return runtime.None, err
	}

	if len(args) == 1 {
		return target.GetInnerText(ctx)
	}

	selector, err := drivers.ToQuerySelector(args[1])

	if err != nil {
		return runtime.None, err
	}

	return target.GetInnerTextBySelector(ctx, selector)
}
