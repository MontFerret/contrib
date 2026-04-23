package input

import (
	"context"

	cdpruntime "github.com/mafredri/cdp/protocol/runtime"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/eval"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/templates"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

var evalRefBySelector = func(
	ctx context.Context,
	exec *eval.Runtime,
	parentID cdpruntime.RemoteObjectID,
	selector drivers.QuerySelector,
) (cdpruntime.RemoteObject, error) {
	return exec.EvalRef(ctx, templates.QuerySelector(parentID, selector))
}

type targetRef struct {
	objectID *cdpruntime.RemoteObjectID
	selector *drivers.QuerySelector
	parentID cdpruntime.RemoteObjectID
}

func directTarget(objectID cdpruntime.RemoteObjectID) targetRef {
	return targetRef{objectID: &objectID}
}

func selectorTarget(parentID cdpruntime.RemoteObjectID, selector drivers.QuerySelector) targetRef {
	return targetRef{
		parentID: parentID,
		selector: &selector,
	}
}

func interactionScrollOptions() drivers.ScrollOptions {
	return drivers.ScrollOptions{
		Behavior: drivers.ScrollBehaviorAuto,
		Block:    drivers.ScrollVerticalAlignmentCenter,
		Inline:   drivers.ScrollHorizontalAlignmentCenter,
	}
}

func (m *Manager) resolveTargetID(
	ctx context.Context,
	target targetRef,
	options drivers.ScrollOptions,
) (cdpruntime.RemoteObjectID, error) {
	if err := target.scroll(ctx, m, options); err != nil {
		return "", err
	}

	return target.resolve(func(parentID cdpruntime.RemoteObjectID, selector drivers.QuerySelector) (cdpruntime.RemoteObjectID, error) {
		return m.querySelectorObjectID(ctx, parentID, selector)
	})
}

func (m *Manager) querySelectorObjectID(
	ctx context.Context,
	parentID cdpruntime.RemoteObjectID,
	selector drivers.QuerySelector,
) (cdpruntime.RemoteObjectID, error) {
	found, err := evalRefBySelector(ctx, m.exec, parentID, selector)
	if err != nil {
		return "", err
	}

	if found.ObjectID == nil {
		return "", runtime.ErrNotFound
	}

	return *found.ObjectID, nil
}

func (target targetRef) scroll(ctx context.Context, manager *Manager, options drivers.ScrollOptions) error {
	if target.objectID != nil {
		return manager.ScrollIntoView(ctx, *target.objectID, options)
	}

	if target.selector == nil {
		return runtime.Error(runtime.ErrMissedArgument, "selector")
	}

	return manager.ScrollIntoViewBySelector(ctx, target.parentID, *target.selector, options)
}

func (target targetRef) resolve(
	resolveSelector func(parentID cdpruntime.RemoteObjectID, selector drivers.QuerySelector) (cdpruntime.RemoteObjectID, error),
) (cdpruntime.RemoteObjectID, error) {
	if target.objectID != nil {
		return *target.objectID, nil
	}

	if target.selector == nil {
		return "", runtime.Error(runtime.ErrMissedArgument, "selector")
	}

	return resolveSelector(target.parentID, *target.selector)
}
