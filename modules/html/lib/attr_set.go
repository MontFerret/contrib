package html

import (
	"context"

	"github.com/MontFerret/contrib/modules/html/drivers/common"
	"github.com/MontFerret/ferret/v2/pkg/runtime"

	"github.com/MontFerret/contrib/modules/html/drivers"
)

// ATTR_SET sets or updates a single or more attribute(s) of a given element.
// @param {HTMLPage | HTMLDocument | HTMLElement} node - Target node.
// @param {String | Object} nameOrObj - Attribute name or an object representing a key-value pair of attributes.
// @param {String} value - If a second parameter is a string value, this parameter represent an attribute value.
func AttributeSet(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 2, runtime.MaxArgs)

	if err != nil {
		return runtime.None, err
	}

	el, err := drivers.ToElement(args[0])

	if err != nil {
		return runtime.None, err
	}

	switch arg1 := args[1].(type) {
	case runtime.String:
		// ATTR_SET(el, name, value)
		err = runtime.ValidateArgs(args, 3, 3)

		if err != nil {
			return runtime.None, err
		}

		switch arg2 := args[2].(type) {
		case runtime.String:
			return runtime.None, el.SetAttribute(ctx, arg1, arg2)
		case *runtime.Object:
			if arg1 == common.AttrNameStyle {
				styles, err := common.SerializeStyles(ctx, arg2)

				if err != nil {
					return runtime.None, err
				}

				return runtime.None, el.SetAttribute(ctx, arg1, styles)
			}

			return runtime.None, el.SetAttribute(ctx, arg1, runtime.NewString(arg2.String()))
		default:
			return runtime.None, runtime.TypeErrorOf(arg1, runtime.TypeString, runtime.TypeMap)
		}
	case *runtime.Object:
		// ATTR_SET(el, values)
		return runtime.None, el.SetAttributes(ctx, arg1)
	default:
		return runtime.None, runtime.TypeErrorOf(arg1, runtime.TypeString, runtime.TypeMap)
	}
}
