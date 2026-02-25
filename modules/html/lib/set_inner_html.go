package html

import (
	"context"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// INNER_HTML_SET sets inner HTML string to a given or matched by CSS selector element
// @param {HTMLPage | HTMLDocument | HTMLElement} node - Target html node.
// @param {String} htmlOrSelector - HTML or CSS selector.
// @param {String} [html] - String of inner HTML.
func SetInnerHTML(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
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

		return runtime.None, el.SetInnerHTML(ctx, runtime.ToString(args[1]))
	}

	selector, err := drivers.ToQuerySelector(args[1])

	if err != nil {
		return runtime.None, err
	}

	err = runtime.ValidateType(args[2], runtime.TypeString)

	if err != nil {
		return runtime.None, err
	}

	innerHTML := runtime.ToString(args[2])

	return runtime.None, el.SetInnerHTMLBySelector(ctx, selector, innerHTML)
}
