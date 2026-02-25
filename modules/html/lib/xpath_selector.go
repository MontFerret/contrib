package html

import (
	"context"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// X returns QuerySelector of XPath kind.
// @param {String} expression - XPath expression.
// @return {Any} - Returns QuerySelector of XPath kind.
func XPathSelector(_ context.Context, expression runtime.Value) (runtime.Value, error) {
	return sdk.NewProxy(drivers.NewXPathSelector(runtime.ToString(expression))), nil
}
