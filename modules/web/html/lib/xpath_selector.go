package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// XPathSelector creates a reusable XPath query selector.
//
// @param expression {String} XPath expression.
// @return {QuerySelector} XPath query selector.
func XPathSelector(_ context.Context, expression runtime.Value) (runtime.Value, error) {
	selector := drivers.NewXPathSelector(runtime.ToString(expression))

	return sdk.NewHostValueWithType(drivers.TypeQuerySelector, selector), nil
}
