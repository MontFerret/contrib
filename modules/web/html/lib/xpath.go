package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// XPath evaluates the XPath expression.
// @param {HTMLPage | HTMLDocument | HTMLElement} node - Target html node.
// @param {String} expression - XPath expression.
// @return {Any} - Returns result of a given XPath expression.
func XPath(ctx context.Context, arg1, arg2 runtime.Value) (runtime.Value, error) {
	target, err := drivers.ToQueryTarget(arg1)

	if err != nil {
		return runtime.None, err
	}

	expr := runtime.ToString(arg2)

	return target.XPath(ctx, expr)
}
