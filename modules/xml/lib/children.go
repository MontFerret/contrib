package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/xml/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Children returns the child nodes for an XML document or element.
//
// @param value {Object} XML document, element, or text node.
// @return {Array<Object>} Child nodes, or an empty array for a text node.
func Children(ctx context.Context, value runtime.Value) (runtime.Value, error) {
	return core.Children(ctx, value)
}
