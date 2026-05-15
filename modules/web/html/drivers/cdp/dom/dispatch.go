package dom

import (
	"context"
	"time"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/input"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

var supportedDispatchEvents = []string{
	drivers.DispatchClickEvent,
	drivers.DispatchDoubleClickEvent,
	drivers.DispatchMouseDownEvent,
	drivers.DispatchMouseUpEvent,
	drivers.DispatchMouseOverEvent,
	drivers.DispatchMouseOutEvent,
	drivers.DispatchMouseMoveEvent,
	drivers.DispatchKeyDownEvent,
	drivers.DispatchKeyUpEvent,
	drivers.DispatchKeyPressEvent,
	drivers.DispatchPressEvent,
	drivers.DispatchTypeEvent,
	drivers.DispatchInputEvent,
	drivers.DispatchChangeEvent,
	drivers.DispatchSubmitEvent,
	drivers.DispatchResetEvent,
	drivers.DispatchFocusEvent,
	drivers.DispatchBlurEvent,
	drivers.DispatchCheckEvent,
	drivers.DispatchUncheckEvent,
	drivers.DispatchToggleEvent,
	drivers.DispatchScrollEvent,
}

func validateDispatchEvent(event runtime.DispatchEvent) error {
	if event.Options != nil && event.Options != runtime.None {
		return runtime.Error(runtime.ErrInvalidOperation, "dispatch options are not supported")
	}

	if drivers.IsDispatchEvent(event.Name.String()) {
		return nil
	}

	return runtime.Errorf(
		runtime.ErrInvalidOperation,
		"unknown dispatch event %q; supported events: %s",
		event.Name.String(),
		supportedDispatchEventNames(),
	)
}

func dispatchHTMLDocument(ctx context.Context, doc *HTMLDocument, event runtime.DispatchEvent) error {
	if err := validateDispatchEvent(event); err != nil {
		return err
	}

	if event.Name == drivers.DispatchScrollEvent {
		return dispatchDocumentScroll(ctx, doc, event.Payload)
	}

	return dispatchHTMLElement(ctx, doc.element, event, true)
}

func dispatchHTMLElement(ctx context.Context, el *HTMLElement, event runtime.DispatchEvent, validated bool) error {
	if !validated {
		if err := validateDispatchEvent(event); err != nil {
			return err
		}
	}

	eventName := event.Name.String()

	switch eventName {
	case drivers.DispatchClickEvent,
		drivers.DispatchDoubleClickEvent,
		drivers.DispatchMouseDownEvent,
		drivers.DispatchMouseUpEvent,
		drivers.DispatchMouseOverEvent,
		drivers.DispatchMouseOutEvent,
		drivers.DispatchMouseMoveEvent:
		params, err := parseDispatchMousePayload(ctx, eventName, event.Payload)
		if err != nil {
			return err
		}

		return el.input.MouseEvent(ctx, el.id, eventName, params)
	case drivers.DispatchKeyDownEvent,
		drivers.DispatchKeyUpEvent,
		drivers.DispatchKeyPressEvent:
		key, err := parseDispatchKeyPayload(ctx, event.Payload)
		if err != nil {
			return err
		}

		return el.input.KeyEvent(ctx, el.id, eventName, key)
	case drivers.DispatchPressEvent:
		params, err := parseDispatchKeyboardPayload(ctx, event.Payload)
		if err != nil {
			return err
		}

		if err := el.Focus(ctx); err != nil {
			return err
		}

		return el.input.Press(ctx, sdk.UnwrapStrings(params.Keys), int(params.Count))
	case drivers.DispatchTypeEvent:
		params, err := parseDispatchTypePayload(ctx, event.Payload)
		if err != nil {
			return err
		}

		return el.input.Type(ctx, el.id, input.TypeParams{
			Text:  params.Text,
			Clear: params.Clear,
			Delay: durationFromRuntimeInt(params.Delay),
		})
	case drivers.DispatchInputEvent:
		payload, err := newDispatchPayload(event.Payload)
		if err != nil {
			return err
		}

		value, err := dispatchRequire(ctx, payload, "value")
		if err != nil {
			return err
		}

		return el.input.InputEvent(ctx, el.id, value)
	case drivers.DispatchChangeEvent:
		payload, err := newDispatchPayload(event.Payload)
		if err != nil {
			return err
		}

		value, hasValue, err := dispatchLookup(ctx, payload, "value")
		if err != nil {
			return err
		}

		return el.input.ChangeEvent(ctx, el.id, value, hasValue && value != runtime.None)
	case drivers.DispatchSubmitEvent:
		return el.input.SubmitEvent(ctx, el.id)
	case drivers.DispatchResetEvent:
		return el.input.ResetEvent(ctx, el.id)
	case drivers.DispatchFocusEvent:
		return el.Focus(ctx)
	case drivers.DispatchBlurEvent:
		return el.Blur(ctx)
	case drivers.DispatchCheckEvent,
		drivers.DispatchUncheckEvent,
		drivers.DispatchToggleEvent:
		return el.input.CheckEvent(ctx, el.id, eventName)
	case drivers.DispatchScrollEvent:
		return dispatchElementScroll(ctx, el, event.Payload)
	default:
		return runtime.Errorf(runtime.ErrInvalidOperation, "unknown dispatch event: %s", event.Name)
	}
}

func durationFromRuntimeInt(value runtime.Int) time.Duration {
	return time.Duration(value) * time.Millisecond
}

func dispatchDocumentScroll(ctx context.Context, doc *HTMLDocument, payload runtime.Value) error {
	params, err := parseDispatchScrollPayload(ctx, payload)
	if err != nil {
		return err
	}

	switch params.Mode {
	case dispatchScrollModeIntoView:
		return doc.element.ScrollIntoView(ctx, params.Options)
	case dispatchScrollModeBy:
		return doc.input.ScrollByDelta(ctx, params.Options)
	default:
		return doc.Scroll(ctx, params.Options)
	}
}

func dispatchElementScroll(ctx context.Context, el *HTMLElement, payload runtime.Value) error {
	params, err := parseDispatchScrollPayload(ctx, payload)
	if err != nil {
		return err
	}

	if params.Mode == dispatchScrollModeIntoView {
		return el.ScrollIntoView(ctx, params.Options)
	}

	return el.input.ElementScroll(ctx, el.id, string(params.Mode), params.Options)
}
