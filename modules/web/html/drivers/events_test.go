package drivers

import "testing"

func TestIsDispatchEvent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want bool
	}{
		{DispatchClickEvent, true},
		{DispatchDoubleClickEvent, true},
		{DispatchMouseDownEvent, true},
		{DispatchMouseUpEvent, true},
		{DispatchMouseOverEvent, true},
		{DispatchMouseOutEvent, true},
		{DispatchMouseMoveEvent, true},
		{DispatchKeyDownEvent, true},
		{DispatchKeyUpEvent, true},
		{DispatchKeyPressEvent, true},
		{DispatchPressEvent, true},
		{DispatchTypeEvent, true},
		{DispatchInputEvent, true},
		{DispatchChangeEvent, true},
		{DispatchSubmitEvent, true},
		{DispatchResetEvent, true},
		{DispatchFocusEvent, true},
		{DispatchBlurEvent, true},
		{DispatchCheckEvent, true},
		{DispatchUncheckEvent, true},
		{DispatchToggleEvent, true},
		{DispatchScrollEvent, true},
		{"Click", false},
		{"network.request_started", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := IsDispatchEvent(tt.name); got != tt.want {
			t.Fatalf("IsDispatchEvent(%q) = %t, want %t", tt.name, got, tt.want)
		}
	}
}
