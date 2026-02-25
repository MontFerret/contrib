package html

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/pkg/errors"

	"github.com/MontFerret/contrib/modules/html/drivers"
)

// SCROLL_ELEMENT scrolls an element on.
// @param {HTMLPage | HTMLDocument | HTMLElement} node - Target html node.
// @param {String} selector - If document is passed, this param must represent an element selector.
// @param {Object} [params] - Scroll params.
// @param {String} [params.behavior="instant"] - Scroll behavior
// @param {String} [params.block="center"] - Scroll vertical alignment.
// @param {String} [params.inline="center"] - Scroll horizontal alignment.
func ScrollInto(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 1, 3)

	if err != nil {
		return runtime.None, err
	}

	var doc drivers.HTMLDocument
	var el drivers.HTMLElement
	var selector drivers.QuerySelector
	var opts drivers.ScrollOptions

	if len(args) == 3 {
		if err = runtime.ValidateType(args[1], runtime.TypeString); err != nil {
			return runtime.None, errors.Wrap(err, "selector")
		}

		if err = runtime.ValidateType(args[2], runtime.TypeObject); err != nil {
			return runtime.None, errors.Wrap(err, "options")
		}

		doc, err = drivers.ToDocument(args[0])

		if err != nil {
			return runtime.None, errors.Wrap(err, "document")
		}

		selector, err = drivers.ToQuerySelector(args[1])

		if err != nil {
			return runtime.None, err
		}

		o, err := toScrollOptions(args[2])

		if err != nil {
			return runtime.None, errors.Wrap(err, "options")
		}

		opts = o
	} else if len(args) == 2 {
		if err = runtime.ValidateType(args[1], runtime.TypeString, runtime.TypeObject); err != nil {
			return runtime.None, err
		}

		switch argv := args[1].(type) {
		case runtime.String:
			doc, err = drivers.ToDocument(args[0])

			if err != nil {
				return runtime.None, errors.Wrap(err, "document")
			}

			selector, err = drivers.ToQuerySelector(argv)

			if err != nil {
				return runtime.None, err
			}
		default:
			el, err = drivers.ToElement(args[0])
			o, err := toScrollOptions(args[1])

			if err != nil {
				return runtime.None, errors.Wrap(err, "options")
			}

			opts = o

		}
	} else {
		el, err = drivers.ToElement(args[0])

		if err != nil {
			return runtime.None, errors.Wrap(err, "element")
		}
	}

	if doc != nil {
		if selector.String() != "" {
			return runtime.True, doc.ScrollBySelector(ctx, selector, opts)
		}

		return runtime.True, doc.GetElement().ScrollIntoView(ctx, opts)
	}

	if el != nil {
		return runtime.True, el.ScrollIntoView(ctx, opts)
	}

	return runtime.None, runtime.TypeError(
		runtime.TypeOf(args[0]),
		drivers.HTMLPageType,
		drivers.HTMLDocumentType,
		drivers.HTMLElementType,
	)
}
