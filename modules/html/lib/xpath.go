package html

import (
	"context"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// XPATH evaluates the XPath expression.
// @param {HTMLPage | HTMLDocument | HTMLElement} node - Target html node.
// @param {String} expression - XPath expression.
// @return {Any} - Returns result of a given XPath expression.
func XPath(ctx context.Context, arg1, arg2 runtime.Value) (runtime.Value, error) {
	element, err := drivers.ToElement(arg1)

	if err != nil {
		return runtime.None, err
	}

	expr := runtime.ToString(arg2)

	return element.XPath(ctx, expr)
}
