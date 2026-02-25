package html

import (
	"context"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// ATTR_QUERY finds a single or more attribute(s) by an query selector.
// @param {HTMLPage | HTMLDocument | HTMLElement} node - Target node.
// @param {String} selector - Query selector.
// @param {String, repeated} attrName - Attr name(s).
// @return {Object} - Key-value pairs of attribute runtime.
func AttributeQuery(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 2, runtime.MaxArgs)

	if err != nil {
		return runtime.None, err
	}

	parent, err := drivers.ToElement(args[0])

	if err != nil {
		return runtime.None, err
	}

	selector, err := drivers.ToQuerySelector(args[1])

	if err != nil {
		return runtime.None, err
	}

	found, err := parent.QuerySelector(ctx, selector)

	if err != nil {
		return runtime.None, err
	}

	el, err := drivers.ToElement(found)

	if err != nil {
		return runtime.None, err
	}

	names := args[2:]
	result := runtime.NewObject()
	attrs, err := el.GetAttributes(ctx)

	if err != nil {
		return runtime.None, err
	}

	for _, n := range names {
		name := runtime.NewString(n.String())
		val, err := attrs.Get(ctx, name)

		if err != nil {
			return runtime.None, err
		}

		if val != runtime.None {
			if err := result.Set(ctx, name, val); err != nil {
				return runtime.None, err
			}
		}
	}

	return result, nil
}
