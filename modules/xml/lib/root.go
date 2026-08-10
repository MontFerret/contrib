package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/xml/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Root returns the root element for a document or the element itself.
//
// @param value {Object} XML document, element, or text node.
// @return {Object|None} Root or element node, or None for a text node.
func Root(ctx context.Context, value runtime.Value) (runtime.Value, error) {
	return core.Root(ctx, value)
}
