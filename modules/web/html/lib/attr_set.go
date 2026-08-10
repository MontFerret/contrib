package lib

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"

	"github.com/MontFerret/contrib/modules/web/html/internal/styleutil"
)

// AttributeSet sets one or more attributes on an HTML root.
//
// @param root {HTMLPage|HTMLDocument|HTMLElement} HTML root.
// @param nameOrAttributes {String|Object} Attribute name or attribute map.
// @param value {String|Object?} Attribute value when a name is supplied.
// @return {None} No value.
func AttributeSet(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 2, runtime.MaxArgs)

	if err != nil {
		return runtime.None, err
	}

	target, err := toRootAttributeTarget(args[0])

	if err != nil {
		return runtime.None, runtime.ArgError(err, 0)
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
			return runtime.None, target.SetAttribute(ctx, arg1, arg2)
		case runtime.Map:
			if arg1 == styleutil.AttributeNameStyle {
				styles, err := styleutil.Serialize(ctx, arg2)

				if err != nil {
					return runtime.None, err
				}

				return runtime.None, target.SetAttribute(ctx, arg1, styles)
			}

			return runtime.None, target.SetAttribute(ctx, arg1, runtime.NewString(arg2.String()))
		default:
			return runtime.None, runtime.ArgError(runtime.TypeErrorOf(arg2, runtime.TypeString, runtime.TypeMap), 2)
		}
	case runtime.Map:
		// ATTR_SET(el, values)
		return runtime.None, target.SetAttributes(ctx, arg1)
	default:
		return runtime.None, runtime.TypeErrorOf(arg1, runtime.TypeString, runtime.TypeMap)
	}
}
