package html

import (
	"context"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// STYLE_REMOVE removes single or more style attribute value(s) of a given element.
// @param {HTMLElement} element - Target html element.
// @param {String, repeated} names - Style name(s).
func StyleRemove(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
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
			return runtime.None, runtime.TypeError(runtime.TypeOf(attr), runtime.TypeString)
		}

		attrsStr = append(attrsStr, str)
	}

	return runtime.None, el.RemoveStyle(ctx, attrsStr...)
}
