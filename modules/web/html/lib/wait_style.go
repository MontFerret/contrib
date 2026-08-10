package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// WaitStyle waits for a style value to appear.
//
// @param root {HTMLPage|HTMLDocument|HTMLElement} HTML root.
// @param selectorOrName {String} Element selector or style name.
// @param nameOrValue {Any} Style name or expected value.
// @param valueOrTimeout {Any?} Expected value or timeout.
// @param timeout {Int?} Wait timeout in milliseconds.
// @return {Boolean} True when the expected state is observed.
func WaitStyle(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	return waitStyleWhen(ctx, args, drivers.WaitEventPresence)
}

// WaitNoStyle waits for a style value to disappear.
//
// @param root {HTMLPage|HTMLDocument|HTMLElement} HTML root.
// @param selectorOrName {String} Element selector or style name.
// @param nameOrValue {Any} Style name or expected value.
// @param valueOrTimeout {Any?} Expected value or timeout.
// @param timeout {Int?} Wait timeout in milliseconds.
// @return {Boolean} True when the expected state is observed.
func WaitNoStyle(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	return waitStyleWhen(ctx, args, drivers.WaitEventAbsence)
}

func waitStyleWhen(ctx context.Context, args []runtime.Value, when drivers.WaitEvent) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 3, 5)

	if err != nil {
		return runtime.None, err
	}

	// document or element
	arg1 := args[0]
	err = runtime.ValidateType(arg1, drivers.HTMLPageType, drivers.HTMLDocumentType, drivers.HTMLElementType)

	if err != nil {
		return runtime.None, err
	}

	timeout := runtime.NewInt(drivers.DefaultWaitTimeout)

	switch arg1.(type) {
	// if a document is passed
	// WAIT_ATTR(doc, selector, attrName, attrValue, timeout)
	case drivers.HTMLPage, drivers.HTMLDocument:
		// revalidate args with more accurate amount
		err := runtime.ValidateArgs(args, 4, 5)

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

		target, err := toRootWaitTarget(arg1)

		if err != nil {
			return runtime.None, err
		}

		name := args[2].(runtime.String)
		value := args[3]

		if len(args) == 5 {
			err = runtime.ValidateType(args[4], runtime.TypeInt)

			if err != nil {
				return runtime.None, err
			}

			timeout = args[4].(runtime.Int)
		}

		ctx, fn := waitTimeout(ctx, timeout)
		defer fn()

		return runtime.True, target.WaitForStyleBySelector(ctx, selector, name, value, when)
	default:
		target, err := toRootWaitTarget(arg1)
		if err != nil {
			return runtime.None, err
		}

		name := args[1].(runtime.String)
		value := args[2]

		if len(args) == 4 {
			err = runtime.ValidateType(args[3], runtime.TypeInt)

			if err != nil {
				return runtime.None, err
			}

			timeout = args[3].(runtime.Int)
		}

		ctx, fn := waitTimeout(ctx, timeout)
		defer fn()

		return runtime.True, target.WaitForStyle(ctx, name, value, when)
	}
}
