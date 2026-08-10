package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// XPath evaluates an XPath expression against an HTML root.
//
// @param root {HTMLPage|HTMLDocument|HTMLElement} HTML root.
// @param expression {String} XPath expression.
// @return {Any} XPath evaluation result.
func XPath(ctx context.Context, arg1, arg2 runtime.Value) (runtime.Value, error) {
	target, err := drivers.ToQueryTarget(arg1)

	if err != nil {
		return runtime.None, err
	}

	expr := runtime.ToString(arg2)

	return target.XPath(ctx, expr)
}
