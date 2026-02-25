package html

import (
	"context"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// INNER_TEXT_SET sets inner text string to a given or matched by CSS selector element
// @param {HTMLPage | HTMLDocument | HTMLElement} node - Target html node.
// @param {String} textOrCssSelector - String of CSS selector.
// @param {String} [text] - String of inner text.
func SetInnerText(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 2, 3)

	if err != nil {
		return runtime.None, err
	}

	el, err := drivers.ToElement(args[0])

	if err != nil {
		return runtime.None, err
	}

	if len(args) == 2 {
		err := runtime.ValidateType(args[1], runtime.TypeString)

		if err != nil {
			return runtime.None, err
		}

		return runtime.None, el.SetInnerText(ctx, runtime.ToString(args[1]))
	}

	err = runtime.ValidateType(args[2], runtime.TypeString)

	if err != nil {
		return runtime.None, err
	}

	selector, err := drivers.ToQuerySelector(args[1])

	if err != nil {
		return runtime.None, err
	}

	innerHTML := runtime.ToString(args[2])

	return runtime.None, el.SetInnerTextBySelector(ctx, selector, innerHTML)
}
