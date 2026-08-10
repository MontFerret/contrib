package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// SetInnerText sets text on a root or selected element.
//
// @param root {HTMLPage|HTMLDocument|HTMLElement} HTML root.
// @param textOrSelector {String} Text content or element selector.
// @param text {String?} Text content when a selector is supplied.
// @return {None} No value.
func SetInnerText(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 2, 3)

	if err != nil {
		return runtime.None, err
	}

	target, err := toRootContentTarget(args[0])

	if err != nil {
		return runtime.None, err
	}

	if len(args) == 2 {
		err := runtime.ValidateType(args[1], runtime.TypeString)

		if err != nil {
			return runtime.None, err
		}

		return runtime.None, target.SetInnerText(ctx, runtime.ToString(args[1]))
	}

	err = runtime.ValidateType(args[2], runtime.TypeString)

	if err != nil {
		return runtime.None, err
	}

	selector, err := drivers.ToQuerySelector(ctx, args[1])

	if err != nil {
		return runtime.None, err
	}

	innerHTML := runtime.ToString(args[2])

	return runtime.None, target.SetInnerTextBySelector(ctx, selector, innerHTML)
}
