package dom

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func (el *HTMLElement) WaitForElement(ctx context.Context, selector drivers.QuerySelector, when drivers.WaitEvent) error {
	return el.wait.WaitForElement(ctx, selector, when)
}

func (el *HTMLElement) WaitForElementAll(ctx context.Context, selector drivers.QuerySelector, when drivers.WaitEvent) error {
	return el.wait.WaitForElementAll(ctx, selector, when)
}

func (el *HTMLElement) WaitForClass(ctx context.Context, class runtime.String, when drivers.WaitEvent) error {
	return el.wait.WaitForClass(ctx, class, when)
}

func (el *HTMLElement) WaitForClassBySelector(ctx context.Context, selector drivers.QuerySelector, class runtime.String, when drivers.WaitEvent) error {
	return el.wait.WaitForClassBySelector(ctx, selector, class, when)
}

func (el *HTMLElement) WaitForClassBySelectorAll(ctx context.Context, selector drivers.QuerySelector, class runtime.String, when drivers.WaitEvent) error {
	return el.wait.WaitForClassBySelectorAll(ctx, selector, class, when)
}

func (el *HTMLElement) WaitForAttribute(
	ctx context.Context,
	name runtime.String,
	value runtime.Value,
	when drivers.WaitEvent,
) error {
	return el.wait.WaitForAttribute(ctx, name, value, when)
}

func (el *HTMLElement) WaitForAttributeBySelector(ctx context.Context, selector drivers.QuerySelector, name runtime.String, value runtime.Value, when drivers.WaitEvent) error {
	return el.wait.WaitForAttributeBySelector(ctx, selector, name, value, when)
}

func (el *HTMLElement) WaitForAttributeBySelectorAll(ctx context.Context, selector drivers.QuerySelector, name runtime.String, value runtime.Value, when drivers.WaitEvent) error {
	return el.wait.WaitForAttributeBySelectorAll(ctx, selector, name, value, when)
}

func (el *HTMLElement) WaitForStyle(ctx context.Context, name runtime.String, value runtime.Value, when drivers.WaitEvent) error {
	return el.wait.WaitForStyle(ctx, name, value, when)
}

func (el *HTMLElement) WaitForStyleBySelector(ctx context.Context, selector drivers.QuerySelector, name runtime.String, value runtime.Value, when drivers.WaitEvent) error {
	return el.wait.WaitForStyleBySelector(ctx, selector, name, value, when)
}

func (el *HTMLElement) WaitForStyleBySelectorAll(ctx context.Context, selector drivers.QuerySelector, name runtime.String, value runtime.Value, when drivers.WaitEvent) error {
	return el.wait.WaitForStyleBySelectorAll(ctx, selector, name, value, when)
}

func (el *HTMLElement) Subscribe(ctx context.Context, subscription runtime.Subscription) (runtime.Stream, error) {
	return subscribeDOMTargetEvents(
		ctx,
		el.client.Runtime,
		el.eval,
		el.id,
		subscription,
	)
}
