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

func (m *Manager) focusTarget(ctx context.Context, target targetRef) error {
	objectID, err := m.resolveTargetID(ctx, target, interactionScrollOptions())
	if err != nil {
		return err
	}

	if err := m.client.DOM.Focus(ctx, dom.NewFocusArgs().SetObjectID(objectID)); err != nil {
		m.logger.Trace().Err(err).Msg("failed focusing on an element")

		return err
	}

	return nil
}

func (m *Manager) blurTarget(ctx context.Context, target targetRef) error {
	switch {
	case target.objectID != nil:
		if err := m.exec.Eval(ctx, templates.Blur(*target.objectID)); err != nil {
			m.logger.Trace().
				Err(err).
				Msg("failed removing focus from an element")

			return err
		}
	case target.selector != nil:
		if err := m.exec.Eval(ctx, templates.BlurBySelector(target.parentID, *target.selector)); err != nil {
			m.logger.Trace().
				Err(err).
				Msg("failed removing focus from an element by selector")

			return err
		}
	default:
		return runtime.Error(runtime.ErrMissedArgument, "selector")
	}

	return nil
}

func (m *Manager) moveMouseTarget(ctx context.Context, target targetRef) error {
	objectID, err := m.resolveTargetID(ctx, target, drivers.ScrollOptions{})
	if err != nil {
		return err
	}

	points, err := GetClickablePointByObjectID(ctx, m.client, objectID)
	if err != nil {
		m.logger.Trace().Err(err).Msg("failed calculating clickable element points")

		return err
	}

	if err := m.mouse.Move(ctx, points.X, points.Y); err != nil {
		m.logger.Trace().Err(err).Msg("failed to move the mouse")

		return err
	}

	return nil
}

func (m *Manager) clickTarget(ctx context.Context, target targetRef, count int) error {
	objectID, err := m.resolveTargetID(ctx, target, interactionScrollOptions())
	if err != nil {
		return err
	}

	points, err := GetClickablePointByObjectID(ctx, m.client, objectID)
	if err != nil {
		m.logger.Trace().Err(err).Msg("failed calculating clickable element points")

		return err
	}

	delay := time.Duration(drivers.DefaultMouseDelay) * time.Millisecond

	if err := m.mouse.ClickWithCount(ctx, points.X, points.Y, delay, count); err != nil {
		m.logger.Trace().Err(err).Msg("failed to click on an element")

		return err
	}

	return nil
}

func (m *Manager) Focus(ctx context.Context, objectID cdpruntime.RemoteObjectID) error {
	m.logger.Trace().
		Str("object_id", string(objectID)).
		Msg("focusing on an element")

	if err := m.focusTarget(ctx, directTarget(objectID)); err != nil {
		return err
	}

	m.logger.Trace().Msg("focused on an element")

	return nil
}

func (m *Manager) FocusBySelector(ctx context.Context, id cdpruntime.RemoteObjectID, selector drivers.QuerySelector) error {
	m.logger.Trace().
		Str("parent_object_id", string(id)).
		Str("selector", selector.String()).
		Msg("focusing on an element by selector")

	if err := m.focusTarget(ctx, selectorTarget(id, selector)); err != nil {
		return err
	}

	m.logger.Trace().Msg("focused on an element")

	return nil
}

func (m *Manager) Blur(ctx context.Context, objectID cdpruntime.RemoteObjectID) error {
	m.logger.Trace().
		Str("object_id", string(objectID)).
		Msg("removing focus from an element")

	if err := m.blurTarget(ctx, directTarget(objectID)); err != nil {
		return err
	}

	m.logger.Trace().Msg("removed focus from an element")

	return nil
}

func (m *Manager) BlurBySelector(ctx context.Context, id cdpruntime.RemoteObjectID, selector drivers.QuerySelector) error {
	m.logger.Trace().
		Str("parent_object_id", string(id)).
		Str("selector", selector.String()).
		Msg("removing focus from an element by selector")

	if err := m.blurTarget(ctx, selectorTarget(id, selector)); err != nil {
		return err
	}

	m.logger.Trace().Msg("removed focus from an element by selector")

	return nil
}

func (m *Manager) MoveMouse(ctx context.Context, objectID cdpruntime.RemoteObjectID) error {
	m.logger.Trace().
		Str("object_id", string(objectID)).
		Msg("starting to move the mouse towards an element")

	if err := m.moveMouseTarget(ctx, directTarget(objectID)); err != nil {
		return err
	}

	m.logger.Trace().Msg("moved the mouse")

	return nil
}

func (m *Manager) MoveMouseBySelector(ctx context.Context, id cdpruntime.RemoteObjectID, selector drivers.QuerySelector) error {
	m.logger.Trace().
		Str("parent_object_id", string(id)).
		Str("selector", selector.String()).
		Msg("starting to move the mouse towards an element by selector")

	if err := m.moveMouseTarget(ctx, selectorTarget(id, selector)); err != nil {
		return err
	}

	m.logger.Trace().Msg("moved the mouse")

	return nil
}

func (m *Manager) MoveMouseByXY(ctx context.Context, xv, yv runtime.Float) error {
	x := float64(xv)
	y := float64(yv)

	m.logger.Trace().
		Float64("x", x).
		Float64("y", y).
		Msg("starting to move the mouse towards an element by given coordinates")

	if err := m.ScrollByXY(ctx, drivers.ScrollOptions{
		Top:  xv,
		Left: yv,
	}); err != nil {
		return err
	}

	if err := m.mouse.Move(ctx, x, y); err != nil {
		m.logger.Trace().Err(err).Msg("failed to move the mouse towards an element by given coordinates")

		return err
	}

	m.logger.Trace().Msg("moved the mouse")

	return nil
}

func (m *Manager) Click(ctx context.Context, objectID cdpruntime.RemoteObjectID, count int) error {
	m.logger.Trace().
		Str("object_id", string(objectID)).
		Msg("starting to click on an element")

	if err := m.clickTarget(ctx, directTarget(objectID), count); err != nil {
		return err
	}

	m.logger.Trace().Msg("clicked on an element")

	return nil
}

func (m *Manager) ClickBySelector(ctx context.Context, id cdpruntime.RemoteObjectID, selector drivers.QuerySelector, count runtime.Int) error {
	m.logger.Trace().
		Str("parent_object_id", string(id)).
		Str("selector", selector.String()).
		Int64("count", int64(count)).
		Msg("clicking on an element by selector")

	if err := m.clickTarget(ctx, selectorTarget(id, selector), int(count)); err != nil {
		return err
	}

	m.logger.Trace().Msg("clicked on an element")

	return nil
}
