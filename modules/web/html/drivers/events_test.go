package drivers

import (
	"reflect"
	"testing"
)

func TestSupportedNetworkEvents(t *testing.T) {
	t.Parallel()

	expected := []string{
		NetworkRequestStartedEvent,
		NetworkResponseReceivedEvent,
		NetworkRequestFinishedEvent,
		NetworkRequestFailedEvent,
		NetworkIdleEvent,
	}

	if got := SupportedNetworkEvents(); !reflect.DeepEqual(got, expected) {
		t.Fatalf("SupportedNetworkEvents() = %#v, want %#v", got, expected)
	}

	events := SupportedNetworkEvents()
	events[0] = "changed"

	if got := SupportedNetworkEvents()[0]; got != NetworkRequestStartedEvent {
		t.Fatalf("SupportedNetworkEvents() returned mutable state, first event = %q", got)
	}
}

func TestSupportedObservableEvents(t *testing.T) {
	t.Parallel()

	expected := []string{
		NavigationEvent,
		RequestEvent,
		ResponseEvent,
		NetworkRequestStartedEvent,
		NetworkResponseReceivedEvent,
		NetworkRequestFinishedEvent,
		NetworkRequestFailedEvent,
		NetworkIdleEvent,
	}

	if got := SupportedObservableEvents(); !reflect.DeepEqual(got, expected) {
		t.Fatalf("SupportedObservableEvents() = %#v, want %#v", got, expected)
	}

	events := SupportedObservableEvents()
	events[0] = "changed"

	if got := SupportedObservableEvents()[0]; got != NavigationEvent {
		t.Fatalf("SupportedObservableEvents() returned mutable state, first event = %q", got)
	}
}

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

func TestIsNetworkEvent(t *testing.T) {
	t.Parallel()

	for _, event := range SupportedNetworkEvents() {
		if !IsNetworkEvent(event) {
			t.Fatalf("IsNetworkEvent(%q) = false, want true", event)
		}
	}

	for _, event := range []string{NavigationEvent, DispatchClickEvent, "network.unknown", ""} {
		if IsNetworkEvent(event) {
			t.Fatalf("IsNetworkEvent(%q) = true, want false", event)
		}
	}
}

func TestIsObservableEvent(t *testing.T) {
	t.Parallel()

	for _, event := range SupportedObservableEvents() {
		if !IsObservableEvent(event) {
			t.Fatalf("IsObservableEvent(%q) = false, want true", event)
		}
	}

	for _, event := range []string{DispatchClickEvent, "network.unknown", ""} {
		if IsObservableEvent(event) {
			t.Fatalf("IsObservableEvent(%q) = true, want false", event)
		}
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
