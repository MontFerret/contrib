package dom

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/events"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/templates"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func (el *HTMLElement) WaitForElement(ctx context.Context, selector drivers.QuerySelector, when drivers.WaitEvent) error {
	task := events.NewEvalWaitTask(
		el.eval,
		templates.WaitForElement(el.id, selector, when),
		events.DefaultPolling,
	)

	_, err := task.Run(ctx)

	return err
}

func (el *HTMLElement) WaitForElementAll(ctx context.Context, selector drivers.QuerySelector, when drivers.WaitEvent) error {
	task := events.NewEvalWaitTask(
		el.eval,
		templates.WaitForElementAll(el.id, selector, when),
		events.DefaultPolling,
	)

	_, err := task.Run(ctx)

	return err
}

func (el *HTMLElement) WaitForClass(ctx context.Context, class runtime.String, when drivers.WaitEvent) error {
	task := events.NewEvalWaitTask(
		el.eval,
		templates.WaitForClass(el.id, class, when),
		events.DefaultPolling,
	)

	_, err := task.Run(ctx)

	return err
}

func (el *HTMLElement) WaitForClassBySelector(ctx context.Context, selector drivers.QuerySelector, class runtime.String, when drivers.WaitEvent) error {
	task := events.NewEvalWaitTask(
		el.eval,
		templates.WaitForClassBySelector(el.id, selector, class, when),
		events.DefaultPolling,
	)

	_, err := task.Run(ctx)

	return err
}

func (el *HTMLElement) WaitForClassBySelectorAll(ctx context.Context, selector drivers.QuerySelector, class runtime.String, when drivers.WaitEvent) error {
	task := events.NewEvalWaitTask(
		el.eval,
		templates.WaitForClassBySelectorAll(el.id, selector, class, when),
		events.DefaultPolling,
	)

	_, err := task.Run(ctx)

	return err
}

func (el *HTMLElement) WaitForAttribute(
	ctx context.Context,
	name runtime.String,
	value runtime.Value,
	when drivers.WaitEvent,
) error {
	task := events.NewEvalWaitTask(
		el.eval,
		templates.WaitForAttribute(el.id, name, value, when),
		events.DefaultPolling,
	)

	_, err := task.Run(ctx)

	return err
}

func (el *HTMLElement) WaitForAttributeBySelector(ctx context.Context, selector drivers.QuerySelector, name runtime.String, value runtime.Value, when drivers.WaitEvent) error {
	task := events.NewEvalWaitTask(
		el.eval,
		templates.WaitForAttributeBySelector(el.id, selector, name, value, when),
		events.DefaultPolling,
	)

	_, err := task.Run(ctx)

	return err
}

func (el *HTMLElement) WaitForAttributeBySelectorAll(ctx context.Context, selector drivers.QuerySelector, name runtime.String, value runtime.Value, when drivers.WaitEvent) error {
	task := events.NewEvalWaitTask(
		el.eval,
		templates.WaitForAttributeBySelectorAll(el.id, selector, name, value, when),
		events.DefaultPolling,
	)

	_, err := task.Run(ctx)

	return err
}

func (el *HTMLElement) WaitForStyle(ctx context.Context, name runtime.String, value runtime.Value, when drivers.WaitEvent) error {
	task := events.NewEvalWaitTask(
		el.eval,
		templates.WaitForStyle(el.id, name, value, when),
		events.DefaultPolling,
	)

	_, err := task.Run(ctx)

	return err
}

func (el *HTMLElement) WaitForStyleBySelector(ctx context.Context, selector drivers.QuerySelector, name runtime.String, value runtime.Value, when drivers.WaitEvent) error {
	task := events.NewEvalWaitTask(
		el.eval,
		templates.WaitForStyleBySelector(el.id, selector, name, value, when),
		events.DefaultPolling,
	)

	_, err := task.Run(ctx)

	return err
}

func (el *HTMLElement) WaitForStyleBySelectorAll(ctx context.Context, selector drivers.QuerySelector, name runtime.String, value runtime.Value, when drivers.WaitEvent) error {
	task := events.NewEvalWaitTask(
		el.eval,
		templates.WaitForStyleBySelectorAll(el.id, selector, name, value, when),
		events.DefaultPolling,
	)

	_, err := task.Run(ctx)

	return err
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

func (el *HTMLElement) Dispatch(ctx context.Context, event runtime.DispatchEvent) error {
	return runtime.Error(runtime.ErrNotImplemented, "HTMLElement.Dispatch")
}
