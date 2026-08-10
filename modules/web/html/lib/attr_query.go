package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// AttributeQuery returns selected attributes from the first matching element.
//
// @param root {HTMLPage|HTMLDocument|HTMLElement} HTML root.
// @param selector {String} Element selector.
// @param names {String...} Attribute names.
// @return {Object} Existing attribute values keyed by name.
func AttributeQuery(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 2, runtime.MaxArgs)

	if err != nil {
		return runtime.None, err
	}

	parent, err := drivers.ToQueryTarget(args[0])

	if err != nil {
		return runtime.None, err
	}

	selector, err := drivers.ToQuerySelector(ctx, args[1])

	if err != nil {
		return runtime.None, err
	}

	found, err := parent.QuerySelector(ctx, selector)

	if err != nil {
		return runtime.None, err
	}

	target, err := drivers.ToAttributeTarget(found)

	if err != nil {
		return runtime.None, err
	}

	names := args[2:]
	result := runtime.NewObject()
	attrs, err := target.GetAttributes(ctx)

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
