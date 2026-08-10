package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Hover moves the mouse over an HTML root or selected element.
//
// @param root {HTMLPage|HTMLDocument|HTMLElement} HTML root.
// @param selector {String?} Element selector.
// @return {Boolean} Whether the target was hovered.
func Hover(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 1, 2)

	if err != nil {
		return runtime.None, err
	}

	target, err := toRootInteractionTarget(args[0])

	if err != nil {
		return runtime.None, err
	}

	if len(args) == 1 {
		return runtime.True, target.Hover(ctx)
	}

	selector, err := drivers.ToQuerySelector(ctx, args[1])

	if err != nil {
		return runtime.None, err
	}

	return runtime.True, target.HoverBySelector(ctx, selector)
}
