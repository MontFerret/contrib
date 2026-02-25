package html

import (
	"context"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// ATTR_REMOVE removes single or more attribute(s) of a given element.
// @param {HTMLPage | HTMLDocument | HTMLElement} node - Target node.
// @param {String, repeated} attrNames - Attribute name(s).
func AttributeRemove(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 2, runtime.MaxArgs)

	if err != nil {
		return runtime.None, err
	}

	el, err := drivers.ToElement(args[0])

	if err != nil {
		return runtime.None, err
	}

	attrs := args[1:]
	attrsStr := make([]runtime.String, 0, len(attrs))

	for _, attr := range attrs {
		str, ok := attr.(runtime.String)

		if !ok {
			return runtime.None, runtime.TypeErrorOf(attr, runtime.TypeString)
		}

		attrsStr = append(attrsStr, str)
	}

	return runtime.None, el.RemoveAttribute(ctx, attrsStr...)
}
