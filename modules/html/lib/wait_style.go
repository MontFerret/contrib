package html

import (
	"context"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// WAIT_STYLE waits until a target style value appears
// @param {HTMLPage | HTMLDocument | HTMLElement} node - Target html node.
// @param {String} styleNameOrSelector - Style name or CSS selector.
// @param {String | Any} valueOrStyleName - Style value or name.
// @param {Any | Int} [valueOrTimeout] - Style value or wait timeout.
// @param {Int} [timeout=5000] - Wait timeout.
func WaitStyle(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	return waitStyleWhen(ctx, args, drivers.WaitEventPresence)
}

// WAIT_NO_STYLE waits until a target style value disappears
// @param {HTMLPage | HTMLDocument | HTMLElement} node - Target html node.
// @param {String} styleNameOrSelector - Style name or CSS selector.
// @param {String | Any} valueOrStyleName - Style value or name.
// @param {Any | Int} [valueOrTimeout] - Style value or wait timeout.
// @param {Int} [timeout=5000] - Wait timeout.
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

		selector, err := drivers.ToQuerySelector(args[1])

		if err != nil {
			return runtime.None, err
		}

		// attr name
		err = runtime.ValidateType(args[2], runtime.TypeString)

		if err != nil {
			return runtime.None, err
		}

		el, err := drivers.ToElement(arg1)

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

		return runtime.True, el.WaitForStyleBySelector(ctx, selector, name, value, when)
	default:
		el := arg1.(drivers.HTMLElement)
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

		return runtime.True, el.WaitForStyle(ctx, name, value, when)
	}
}
