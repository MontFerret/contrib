package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Unhover moves the mouse outside an HTML root or selected element.
//
// @param root {HTMLPage|HTMLDocument|HTMLElement} HTML root.
// @param selector {String?} Element selector.
// @return {Boolean} Whether the mouse was moved away.
func Unhover(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 1, 2)

	if err != nil {
		return runtime.None, err
	}

	target, err := toRootInteractionTarget(args[0])

	if err != nil {
		return runtime.None, err
	}

	if len(args) == 1 {
		return runtime.True, target.Unhover(ctx)
	}

	selector, err := drivers.ToQuerySelector(ctx, args[1])

	if err != nil {
		return runtime.None, err
	}

	return runtime.True, target.UnhoverBySelector(ctx, selector)
}
