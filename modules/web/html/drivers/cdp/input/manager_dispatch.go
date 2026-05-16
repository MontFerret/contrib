package input

import (
	"context"
	"time"

	cdpruntime "github.com/mafredri/cdp/protocol/runtime"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/templates"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func (m *Manager) MouseEvent(ctx context.Context, objectID cdpruntime.RemoteObjectID, event string, params MouseEventParams) error {
	objectID, err := m.resolveTargetID(ctx, directTarget(objectID), interactionScrollOptions())
	if err != nil {
		return err
	}

	point, err := GetElementPointByObjectID(ctx, m.client, objectID, params.X, params.Y)
	if err != nil {
		m.logger.Trace().Err(err).Msg("failed calculating dispatch mouse point")

		return err
	}

	button, err := toProtocolMouseButton(params.Button)
	if err != nil {
		return err
	}

	count := params.Count
	if count <= 0 {
		count = 1
	}

	delay := time.Duration(drivers.DefaultMouseDelay) * time.Millisecond

	switch event {
	case drivers.DispatchClickEvent:
		return m.mouse.ClickWithButton(ctx, point.X, point.Y, delay, button, count)
	case drivers.DispatchDoubleClickEvent:
		if params.Count <= 0 {
			count = 2
		}

		return m.mouse.ClickWithButton(ctx, point.X, point.Y, delay, button, count)
	case drivers.DispatchMouseDownEvent:
		if err := m.mouse.Move(ctx, point.X, point.Y); err != nil {
			return err
		}

		return m.mouse.DownWithCount(ctx, button, count)
	case drivers.DispatchMouseUpEvent:
		if err := m.mouse.Move(ctx, point.X, point.Y); err != nil {
			return err
		}

		return m.mouse.UpWithCount(ctx, button, count)
	case drivers.DispatchMouseMoveEvent, drivers.DispatchMouseOverEvent:
		return m.mouse.Move(ctx, point.X, point.Y)
	case drivers.DispatchMouseOutEvent:
		if err := m.mouse.Move(ctx, point.X, point.Y); err != nil {
			return err
		}

		return m.mouse.Move(ctx, 0, 0)
	default:
		return runtime.Errorf(runtime.ErrInvalidOperation, "unsupported mouse event: %s", event)
	}
}

func (m *Manager) KeyEvent(ctx context.Context, objectID cdpruntime.RemoteObjectID, event, key string) error {
	if err := m.focusTarget(ctx, directTarget(objectID)); err != nil {
		return err
	}

	switch event {
	case drivers.DispatchKeyDownEvent:
		return m.keyboard.Down(ctx, key)
	case drivers.DispatchKeyUpEvent:
		return m.keyboard.Up(ctx, key)
	case drivers.DispatchKeyPressEvent:
		return m.keyboard.Char(ctx, key)
	default:
		return runtime.Errorf(runtime.ErrInvalidOperation, "unsupported keyboard event: %s", event)
	}
}

func (m *Manager) InputEvent(ctx context.Context, objectID cdpruntime.RemoteObjectID, value runtime.Value) error {
	return m.exec.Eval(ctx, templates.DispatchInput(objectID, value))
}

func (m *Manager) ChangeEvent(ctx context.Context, objectID cdpruntime.RemoteObjectID, value runtime.Value, hasValue bool) error {
	return m.exec.Eval(ctx, templates.DispatchChange(objectID, value, hasValue))
}

func (m *Manager) CheckEvent(ctx context.Context, objectID cdpruntime.RemoteObjectID, action string) error {
	return m.exec.Eval(ctx, templates.DispatchCheck(objectID, action))
}

func (m *Manager) SubmitEvent(ctx context.Context, objectID cdpruntime.RemoteObjectID) error {
	return m.exec.Eval(ctx, templates.DispatchSubmit(objectID))
}

func (m *Manager) ResetEvent(ctx context.Context, objectID cdpruntime.RemoteObjectID) error {
	return m.exec.Eval(ctx, templates.DispatchReset(objectID))
}

func (m *Manager) ElementScroll(ctx context.Context, objectID cdpruntime.RemoteObjectID, mode string, options drivers.ScrollOptions) error {
	return m.exec.Eval(ctx, templates.ElementScroll(objectID, mode, options))
}
