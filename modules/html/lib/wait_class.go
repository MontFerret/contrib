package html

import (
	"context"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// WAIT_CLASS waits for a class to appear on a given element.
// Stops the execution until the navigation ends or operation times out.
// @param {HTMLPage | HTMLDocument | HTMLElement} node - Target html node.
// @param {String} selectorOrClass - If document is passed, this param must represent an element selector. Otherwise target class.
// @param {String | Int} [classOrTimeout] - If document is passed, this param must represent target class name. Otherwise timeout.
// @param {Int} [timeout] - If document is passed, this param must represent timeout. Otherwise not passed.
func WaitClass(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	return waitClassWhen(ctx, args, drivers.WaitEventPresence)
}

// WAIT_NO_CLASS waits for a class to disappear on a given element.
// Stops the execution until the navigation ends or operation times out.
// @param {HTMLPage | HTMLDocument | HTMLElement} node - Target html node.
// @param {String} selectorOrClass - If document is passed, this param must represent an element selector. Otherwise target class.
// @param {String | Int} [classOrTimeout] - If document is passed, this param must represent target class name. Otherwise timeout.
// @param {Int} [timeout] - If document is passed, this param must represent timeout. Otherwise not passed.
func WaitNoClass(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	return waitClassWhen(ctx, args, drivers.WaitEventAbsence)
}

func waitClassWhen(ctx context.Context, args []runtime.Value, when drivers.WaitEvent) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 2, 4)

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
	case drivers.HTMLPage, drivers.HTMLDocument:
		// revalidate args with more accurate amount
		err := runtime.ValidateArgs(args, 3, 4)

		if err != nil {
			return runtime.None, err
		}

		selector, err := drivers.ToQuerySelector(args[1])

		if err != nil {
			return runtime.None, err
		}

		// class
		err = runtime.ValidateType(args[2], runtime.TypeString)

		if err != nil {
			return runtime.None, err
		}

		el, err := drivers.ToElement(arg1)

		if err != nil {
			return runtime.None, err
		}

		class := args[2].(runtime.String)

		if len(args) == 4 {
			err = runtime.ValidateType(args[3], runtime.TypeInt)

			if err != nil {
				return runtime.None, err
			}

			timeout = args[3].(runtime.Int)
		}

		ctx, fn := waitTimeout(ctx, timeout)
		defer fn()

		return runtime.True, el.WaitForClassBySelector(ctx, selector, class, when)
	default:
		el := arg1.(drivers.HTMLElement)
		class := args[1].(runtime.String)

		if len(args) == 3 {
			err = runtime.ValidateType(args[2], runtime.TypeInt)

			if err != nil {
				return runtime.None, err
			}

			timeout = args[2].(runtime.Int)
		}

		ctx, fn := waitTimeout(ctx, timeout)
		defer fn()

		return runtime.True, el.WaitForClass(ctx, class, when)
	}
}
