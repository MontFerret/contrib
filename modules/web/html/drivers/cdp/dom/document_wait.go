package dom

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/events"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/templates"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func (doc *HTMLDocument) WaitForElement(ctx context.Context, selector drivers.QuerySelector, when drivers.WaitEvent) error {
	task := events.NewEvalWaitTask(
		doc.eval,
		templates.WaitForElement(doc.element.id, selector, when),
		events.DefaultPolling,
	)

	_, err := task.Run(ctx)

	return err
}

func (doc *HTMLDocument) WaitForClassBySelector(ctx context.Context, selector drivers.QuerySelector, class runtime.String, when drivers.WaitEvent) error {
	task := events.NewEvalWaitTask(
		doc.eval,
		templates.WaitForClassBySelector(doc.element.id, selector, class, when),
		events.DefaultPolling,
	)

	_, err := task.Run(ctx)

	return err
}

func (doc *HTMLDocument) WaitForClassBySelectorAll(ctx context.Context, selector drivers.QuerySelector, class runtime.String, when drivers.WaitEvent) error {
	task := events.NewEvalWaitTask(
		doc.eval,
		templates.WaitForClassBySelectorAll(doc.element.id, selector, class, when),
		events.DefaultPolling,
	)

	_, err := task.Run(ctx)

	return err
}

func (doc *HTMLDocument) WaitForAttributeBySelector(
	ctx context.Context,
	selector drivers.QuerySelector,
	name,
	value runtime.String,
	when drivers.WaitEvent,
) error {
	task := events.NewEvalWaitTask(
		doc.eval,
		templates.WaitForAttributeBySelector(doc.element.id, selector, name, value, when),
		events.DefaultPolling,
	)

	_, err := task.Run(ctx)

	return err
}

func (doc *HTMLDocument) WaitForAttributeBySelectorAll(
	ctx context.Context,
	selector drivers.QuerySelector,
	name,
	value runtime.String,
	when drivers.WaitEvent,
) error {
	task := events.NewEvalWaitTask(
		doc.eval,
		templates.WaitForAttributeBySelectorAll(doc.element.id, selector, name, value, when),
		events.DefaultPolling,
	)

	_, err := task.Run(ctx)

	return err
}

func (doc *HTMLDocument) WaitForStyleBySelector(ctx context.Context, selector drivers.QuerySelector, name, value runtime.String, when drivers.WaitEvent) error {
	task := events.NewEvalWaitTask(
		doc.eval,
		templates.WaitForStyleBySelector(doc.element.id, selector, name, value, when),
		events.DefaultPolling,
	)

	_, err := task.Run(ctx)

	return err
}

func (doc *HTMLDocument) WaitForStyleBySelectorAll(ctx context.Context, selector drivers.QuerySelector, name, value runtime.String, when drivers.WaitEvent) error {
	task := events.NewEvalWaitTask(
		doc.eval,
		templates.WaitForStyleBySelectorAll(doc.element.id, selector, name, value, when),
		events.DefaultPolling,
	)

	_, err := task.Run(ctx)

	return err
}
