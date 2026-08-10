package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/xml/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Attr returns the named attribute value for an XML element-like node.
//
// @param value {Object} XML document, element, or text node.
// @param name {String} Attribute name.
// @return {String|None} Attribute value or None.
func Attr(ctx context.Context, value, name runtime.Value) (runtime.Value, error) {
	attrName, ok := name.(runtime.String)
	if !ok {
		return nil, runtime.TypeErrorOf(name, runtime.TypeString)
	}

	return core.Attr(ctx, value, attrName)
}
