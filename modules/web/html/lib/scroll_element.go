package lib

import (
	"context"

	"github.com/pkg/errors"

	"github.com/MontFerret/ferret/v2/pkg/runtime"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
)

// ScrollInto scrolls an element into view.
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

	var doc drivers.DocumentViewportTarget
	var target drivers.InteractionTarget
	var selector drivers.QuerySelector
	var opts drivers.ScrollOptions

	if len(args) == 3 {
		if err = runtime.ValidateType(args[1], runtime.TypeString); err != nil {
			return runtime.None, errors.Wrap(err, "selector")
		}

		if err = runtime.ValidateType(args[2], runtime.TypeObject); err != nil {
			return runtime.None, errors.Wrap(err, "options")
		}

		doc, err = drivers.ToDocumentViewportTarget(args[0])

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
			doc, err = drivers.ToDocumentViewportTarget(args[0])

			if err != nil {
				return runtime.None, errors.Wrap(err, "document")
			}

			selector, err = drivers.ToQuerySelector(argv)

			if err != nil {
				return runtime.None, err
			}
		default:
			target, err = toRootInteractionTarget(args[0])
			if err != nil {
				return runtime.None, errors.Wrap(err, "element")
			}

			o, err := toScrollOptions(args[1])

			if err != nil {
				return runtime.None, errors.Wrap(err, "options")
			}

			opts = o

		}
	} else {
		target, err = toRootInteractionTarget(args[0])

		if err != nil {
			return runtime.None, errors.Wrap(err, "element")
		}
	}

	if doc != nil {
		if selector.String() != "" {
			return runtime.True, doc.ScrollBySelector(ctx, selector, opts)
		}

		target, err = toRootInteractionTarget(args[0])
		if err != nil {
			return runtime.None, errors.Wrap(err, "element")
		}

		return runtime.True, target.ScrollIntoView(ctx, opts)
	}

	if target != nil {
		return runtime.True, target.ScrollIntoView(ctx, opts)
	}

	return runtime.None, runtime.TypeError(
		runtime.TypeOf(args[0]),
		drivers.HTMLPageType,
		drivers.HTMLDocumentType,
		drivers.HTMLElementType,
	)
}
