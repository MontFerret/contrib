package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// WaitAttributeAll waits for an attribute value on all matching elements.
//
// @param root {HTMLPage|HTMLDocument|HTMLElement} HTML root.
// @param selector {String} Element selector.
// @param name {String} Attribute name.
// @param value {Any} Expected value.
// @param timeout {Int?} Wait timeout in milliseconds.
// @return {Boolean} True when the expected state is observed on every match.
func WaitAttributeAll(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	return waitAttributeAllWhen(ctx, args, drivers.WaitEventPresence)
}

// WaitNoAttributeAll waits for an attribute value to disappear from all matching elements.
//
// @param root {HTMLPage|HTMLDocument|HTMLElement} HTML root.
// @param selector {String} Element selector.
// @param name {String} Attribute name.
// @param value {Any} Expected value.
// @param timeout {Int?} Wait timeout in milliseconds.
// @return {Boolean} True when the expected state is observed on every match.
func WaitNoAttributeAll(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	return waitAttributeAllWhen(ctx, args, drivers.WaitEventAbsence)
}

func waitAttributeAllWhen(ctx context.Context, args []runtime.Value, when drivers.WaitEvent) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 4, 5)

	if err != nil {
		return runtime.None, err
	}

	target, err := toRootWaitTarget(args[0])

	if err != nil {
		return runtime.None, err
	}

	selector, err := drivers.ToQuerySelector(ctx, args[1])

	if err != nil {
		return runtime.None, err
	}

	// attr name
	err = runtime.ValidateType(args[2], runtime.TypeString)

	if err != nil {
		return runtime.None, err
	}

	name := args[2].(runtime.String)
	value := args[3]
	timeout := runtime.NewInt(drivers.DefaultWaitTimeout)

	if len(args) == 5 {
		err = runtime.ValidateType(args[4], runtime.TypeInt)

		if err != nil {
			return runtime.None, err
		}

		timeout = args[4].(runtime.Int)
	}

	ctx, fn := waitTimeout(ctx, timeout)
	defer fn()

	return runtime.True, target.WaitForAttributeBySelectorAll(ctx, selector, name, value, when)
}
