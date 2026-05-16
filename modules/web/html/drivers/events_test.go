package drivers

import (
	"reflect"
	"testing"
)

func TestSupportedDispatchEvents(t *testing.T) {
	t.Parallel()

	expected := []string{
		DispatchClickEvent,
		DispatchDoubleClickEvent,
		DispatchMouseDownEvent,
		DispatchMouseUpEvent,
		DispatchMouseOverEvent,
		DispatchMouseOutEvent,
		DispatchMouseMoveEvent,
		DispatchKeyDownEvent,
		DispatchKeyUpEvent,
		DispatchKeyPressEvent,
		DispatchPressEvent,
		DispatchTypeEvent,
		DispatchInputEvent,
		DispatchChangeEvent,
		DispatchSubmitEvent,
		DispatchResetEvent,
		DispatchFocusEvent,
		DispatchBlurEvent,
		DispatchCheckEvent,
		DispatchUncheckEvent,
		DispatchToggleEvent,
		DispatchScrollEvent,
	}

	if got := SupportedDispatchEvents(); !reflect.DeepEqual(got, expected) {
		t.Fatalf("SupportedDispatchEvents() = %#v, want %#v", got, expected)
	}

	events := SupportedDispatchEvents()
	events[0] = "changed"

	if got := SupportedDispatchEvents()[0]; got != DispatchClickEvent {
		t.Fatalf("SupportedDispatchEvents() returned mutable state, first event = %q", got)
	}
}

func TestIsDispatchEvent(t *testing.T) {
	t.Parallel()

	for _, event := range SupportedDispatchEvents() {
		if !IsDispatchEvent(event) {
			t.Fatalf("IsDispatchEvent(%q) = false, want true", event)
		}
	}

	for _, event := range []string{"Click", NetworkRequestStartedEvent, ""} {
		if IsDispatchEvent(event) {
			t.Fatalf("IsDispatchEvent(%q) = true, want false", event)
		}
	}
}
