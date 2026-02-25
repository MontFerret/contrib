package html

import (
	"context"
	"errors"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// ATTR_GET gets single or more attribute(s) of a given element.
// @param {HTMLPage | HTMLDocument | HTMLElement} node - Target node.
// @param {String, repeated} attrNames - Attribute name(s).
// @return {Object} - Key-value pairs of attribute runtime.
func AttributeGet(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 2, runtime.MaxArgs)

	if err != nil {
		return runtime.None, err
	}

	el, err := drivers.ToElement(args[0])

	if err != nil {
		return runtime.None, err
	}

	names := args[1:]
	result := runtime.NewObject()
	attrs, err := el.GetAttributes(ctx)

	if err != nil {
		return runtime.None, err
	}

	for _, n := range names {
		name := runtime.NewString(n.String())
		val, err := attrs.Get(ctx, name)

		if err != nil && !errors.Is(err, drivers.ErrNotFound) && !errors.Is(err, runtime.ErrNotFound) {
			return runtime.None, err
		}

		if err == nil {
			_ = result.Set(ctx, name, val)
		}
	}

	return result, nil
}
