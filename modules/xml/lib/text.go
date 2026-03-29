package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/xml/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Text returns the descendant text content for an XML node.
// @param {Object} value - XML document, element, or text node.
// @return {String} - Concatenated text content.
func Text(ctx context.Context, value runtime.Value) (runtime.Value, error) {
	return core.Text(ctx, value)
}
