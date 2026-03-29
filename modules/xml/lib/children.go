package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/xml/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Children returns the child nodes for an XML document or element.
// @param {Object} value - XML document, element, or text node.
// @return {Any[]} - Child nodes or an empty array for text nodes.
func Children(ctx context.Context, value runtime.Value) (runtime.Value, error) {
	return core.Children(ctx, value)
}
