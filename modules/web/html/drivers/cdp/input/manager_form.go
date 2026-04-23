package input

import (
	"context"
	"time"

	"github.com/mafredri/cdp/protocol/dom"
	cdpruntime "github.com/mafredri/cdp/protocol/runtime"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/templates"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func (m *Manager) typeTarget(ctx context.Context, target targetRef, params TypeParams) error {
	objectID, err := m.resolveTargetID(ctx, target, interactionScrollOptions())
	if err != nil {
		return err
	}

	if err := m.client.DOM.Focus(ctx, dom.NewFocusArgs().SetObjectID(objectID)); err != nil {
		m.logger.Trace().Msg("failed to focus on an element")

		return err
	}

	if params.Clear {
		points, err := GetClickablePointByObjectID(ctx, m.client, objectID)
		if err != nil {
			m.logger.Trace().Err(err).Msg("failed calculating clickable element points")

			return err
		}

		if err := m.ClearByXY(ctx, points); err != nil {
			return err
		}
	}

	d := runtime.NumberLowerBoundary(float64(params.Delay))
	beforeTypeDelay := time.Duration(d)

	m.logger.Trace().Float64("delay", d).Msg("calculated pause delay")

	time.Sleep(beforeTypeDelay)

	if err := m.keyboard.Type(ctx, params.Text, params.Delay); err != nil {
		m.logger.Trace().Err(err).Msg("failed to type text")

		return err
	}

	return nil
}

func (m *Manager) clearTarget(ctx context.Context, target targetRef) error {
	objectID, err := m.resolveTargetID(ctx, target, interactionScrollOptions())
	if err != nil {
		return err
	}

	points, err := GetClickablePointByObjectID(ctx, m.client, objectID)
	if err != nil {
		m.logger.Trace().Err(err).Msg("failed calculating clickable element points")

		return err
	}

	if err := m.client.DOM.Focus(ctx, dom.NewFocusArgs().SetObjectID(objectID)); err != nil {
		m.logger.Trace().Err(err).Msg("failed to focus on an element")

		return err
	}

	if err := m.ClearByXY(ctx, points); err != nil {
		m.logger.Trace().Err(err).Msg("failed to clear element")

		return err
	}

	return nil
}

func (m *Manager) selectTarget(
	ctx context.Context,
	target targetRef,
	value runtime.List,
) (runtime.List, error) {
	if err := m.focusTarget(ctx, target); err != nil {
		return runtime.NewArray(0), err
	}

	m.logger.Trace().Msg("selecting values")
	m.logger.Trace().Msg("evaluating a JS function")

	var (
		res any
		err error
	)

	switch {
	case target.objectID != nil:
		res, err = m.exec.EvalValue(ctx, templates.Select(*target.objectID, value))
	case target.selector != nil:
		res, err = m.exec.EvalValue(ctx, templates.SelectBySelector(target.parentID, *target.selector, value))
	default:
		return runtime.NewArray(0), runtime.Error(runtime.ErrMissedArgument, "selector")
	}

	if err != nil {
		m.logger.Trace().Err(err).Msg("failed to evaluate a JS function")

		return runtime.NewArray(0), err
	}

	m.logger.Trace().Msg("validating JS result")

	arr, ok := res.(runtime.List)
	if !ok {
		m.logger.Trace().Msg("JS result validation failed")

		return runtime.NewArray(0), runtime.ErrUnexpected
	}

	m.logger.Trace().Msg("selected values")

	return arr, nil
}

func (m *Manager) Type(ctx context.Context, objectID cdpruntime.RemoteObjectID, params TypeParams) error {
	m.logger.Trace().
		Str("object_id", string(objectID)).
		Msg("starting to type text")

	if err := m.typeTarget(ctx, directTarget(objectID), params); err != nil {
		return err
	}

	m.logger.Trace().Msg("typed text")

	return nil
}

func (m *Manager) TypeBySelector(ctx context.Context, id cdpruntime.RemoteObjectID, selector drivers.QuerySelector, params TypeParams) error {
	m.logger.Trace().
		Str("parent_object_id", string(id)).
		Str("selector", selector.String()).
		Msg("starting to type text by selector")

	if err := m.typeTarget(ctx, selectorTarget(id, selector), params); err != nil {
		return err
	}

	m.logger.Trace().Msg("typed text")

	return nil
}

func (m *Manager) Clear(ctx context.Context, objectID cdpruntime.RemoteObjectID) error {
	m.logger.Trace().
		Str("object_id", string(objectID)).
		Msg("starting to clear element")

	if err := m.clearTarget(ctx, directTarget(objectID)); err != nil {
		return err
	}

	m.logger.Trace().Msg("cleared element")

	return nil
}

func (m *Manager) ClearBySelector(ctx context.Context, id cdpruntime.RemoteObjectID, selector drivers.QuerySelector) error {
	m.logger.Trace().
		Str("parent_object_id", string(id)).
		Str("selector", selector.String()).
		Msg("starting to clear element by selector")

	if err := m.clearTarget(ctx, selectorTarget(id, selector)); err != nil {
		return err
	}

	m.logger.Trace().Msg("cleared element")

	return nil
}

func (m *Manager) ClearByXY(ctx context.Context, points Quad) error {
	m.logger.Trace().
		Float64("x", points.X).
		Float64("y", points.Y).
		Msg("starting to clear element by coordinates")

	delay := time.Duration(drivers.DefaultMouseDelay) * time.Millisecond

	m.logger.Trace().Dur("delay", delay).Msg("clicking mouse button to select text")

	err := m.mouse.ClickWithCount(ctx, points.X, points.Y, delay, 3)
	if err != nil {
		m.logger.Trace().Err(err).Msg("failed to click mouse button")

		return err
	}

	delay = time.Duration(drivers.DefaultKeyboardDelay) * time.Millisecond

	m.logger.Trace().Dur("delay", delay).Msg("pressing 'Backspace'")

	if err := m.keyboard.Press(ctx, []string{"Backspace"}, 1, delay); err != nil {
		m.logger.Trace().Err(err).Msg("failed to press 'Backspace'")

		return err
	}

	return nil
}

func (m *Manager) Press(ctx context.Context, keys []string, count int) error {
	delay := time.Duration(drivers.DefaultKeyboardDelay) * time.Millisecond

	m.logger.Trace().
		Strs("keys", keys).
		Int("count", count).
		Dur("delay", delay).
		Msg("pressing keyboard keys")

	if err := m.keyboard.Press(ctx, keys, count, delay); err != nil {
		m.logger.Trace().Err(err).Msg("failed to press keyboard keys")

		return err
	}

	return nil
}

func (m *Manager) PressBySelector(ctx context.Context, id cdpruntime.RemoteObjectID, selector drivers.QuerySelector, keys []string, count int) error {
	m.logger.Trace().
		Str("parent_object_id", string(id)).
		Str("selector", selector.String()).
		Strs("keys", keys).
		Int("count", count).
		Msg("starting to press keyboard keys by selector")

	if err := m.focusTarget(ctx, selectorTarget(id, selector)); err != nil {
		return err
	}

	return m.Press(ctx, keys, count)
}

func (m *Manager) Select(ctx context.Context, id cdpruntime.RemoteObjectID, value runtime.List) (runtime.List, error) {
	m.logger.Trace().
		Str("object_id", string(id)).
		Msg("starting to select values")

	return m.selectTarget(ctx, directTarget(id), value)
}

func (m *Manager) SelectBySelector(ctx context.Context, id cdpruntime.RemoteObjectID, selector drivers.QuerySelector, value runtime.List) (runtime.List, error) {
	m.logger.Trace().
		Str("parent_object_id", string(id)).
		Str("selector", selector.String()).
		Msg("starting to select values by selector")

	return m.selectTarget(ctx, selectorTarget(id, selector), value)
}
