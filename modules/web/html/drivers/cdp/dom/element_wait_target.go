package dom

import (
	"context"

	cdpruntime "github.com/mafredri/cdp/protocol/runtime"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/eval"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/events"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/templates"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type elementWait struct {
	eval *eval.Runtime
	id   cdpruntime.RemoteObjectID
}

func newElementWait(exec *eval.Runtime, id cdpruntime.RemoteObjectID) *elementWait {
	return &elementWait{
		eval: exec,
		id:   id,
	}
}

func (wait *elementWait) WaitForElement(ctx context.Context, selector drivers.QuerySelector, when drivers.WaitEvent) error {
	return wait.run(ctx, templates.WaitForElement(wait.id, selector, when))
}

func (wait *elementWait) WaitForElementAll(ctx context.Context, selector drivers.QuerySelector, when drivers.WaitEvent) error {
	return wait.run(ctx, templates.WaitForElementAll(wait.id, selector, when))
}

func (wait *elementWait) WaitForClass(ctx context.Context, class runtime.String, when drivers.WaitEvent) error {
	return wait.run(ctx, templates.WaitForClass(wait.id, class, when))
}

func (wait *elementWait) WaitForClassBySelector(ctx context.Context, selector drivers.QuerySelector, class runtime.String, when drivers.WaitEvent) error {
	return wait.run(ctx, templates.WaitForClassBySelector(wait.id, selector, class, when))
}

func (wait *elementWait) WaitForClassBySelectorAll(ctx context.Context, selector drivers.QuerySelector, class runtime.String, when drivers.WaitEvent) error {
	return wait.run(ctx, templates.WaitForClassBySelectorAll(wait.id, selector, class, when))
}

func (wait *elementWait) WaitForAttribute(
	ctx context.Context,
	name runtime.String,
	value runtime.Value,
	when drivers.WaitEvent,
) error {
	return wait.run(ctx, templates.WaitForAttribute(wait.id, name, value, when))
}

func (wait *elementWait) WaitForAttributeBySelector(ctx context.Context, selector drivers.QuerySelector, name runtime.String, value runtime.Value, when drivers.WaitEvent) error {
	return wait.run(ctx, templates.WaitForAttributeBySelector(wait.id, selector, name, value, when))
}

func (wait *elementWait) WaitForAttributeBySelectorAll(ctx context.Context, selector drivers.QuerySelector, name runtime.String, value runtime.Value, when drivers.WaitEvent) error {
	return wait.run(ctx, templates.WaitForAttributeBySelectorAll(wait.id, selector, name, value, when))
}

func (wait *elementWait) WaitForStyle(ctx context.Context, name runtime.String, value runtime.Value, when drivers.WaitEvent) error {
	return wait.run(ctx, templates.WaitForStyle(wait.id, name, value, when))
}

func (wait *elementWait) WaitForStyleBySelector(ctx context.Context, selector drivers.QuerySelector, name runtime.String, value runtime.Value, when drivers.WaitEvent) error {
	return wait.run(ctx, templates.WaitForStyleBySelector(wait.id, selector, name, value, when))
}

func (wait *elementWait) WaitForStyleBySelectorAll(ctx context.Context, selector drivers.QuerySelector, name runtime.String, value runtime.Value, when drivers.WaitEvent) error {
	return wait.run(ctx, templates.WaitForStyleBySelectorAll(wait.id, selector, name, value, when))
}

func (wait *elementWait) run(ctx context.Context, fn *eval.Function) error {
	task := events.NewEvalWaitTask(
		wait.eval,
		fn,
		events.DefaultPolling,
	)

	_, err := task.Run(ctx)

	return err
}
