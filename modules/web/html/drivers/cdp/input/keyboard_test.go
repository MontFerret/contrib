package input

import "testing"

func TestKeyboardCreatePressEventEnter(t *testing.T) {
	keyboard := &Keyboard{}

	down := keyboard.createPressEvent("keyDown", "Enter")

	assertStringPointer(t, down.Code, "Enter")
	assertStringPointer(t, down.Key, "Enter")
	assertIntPointer(t, down.Modifiers, 0)
	assertIntPointer(t, down.WindowsVirtualKeyCode, 13)
	assertIntPointer(t, down.NativeVirtualKeyCode, 13)
	assertIntPointer(t, down.Location, 0)
	assertStringPointer(t, down.Text, "\r")
	assertStringPointer(t, down.UnmodifiedText, "\r")

	up := keyboard.createPressEvent("keyUp", "Enter")

	assertStringPointer(t, up.Code, "Enter")
	assertStringPointer(t, up.Key, "Enter")
	assertIntPointer(t, up.Modifiers, 0)
	assertIntPointer(t, up.WindowsVirtualKeyCode, 13)
	assertIntPointer(t, up.NativeVirtualKeyCode, 13)
	assertIntPointer(t, up.Location, 0)
	assertNilStringPointer(t, up.Text)
	assertNilStringPointer(t, up.UnmodifiedText)
}

func TestKeyboardCreatePressEventBackspace(t *testing.T) {
	keyboard := &Keyboard{}

	event := keyboard.createPressEvent("keyDown", "Backspace")

	assertStringPointer(t, event.Code, "Backspace")
	assertStringPointer(t, event.Key, "Backspace")
	assertIntPointer(t, event.WindowsVirtualKeyCode, 8)
	assertIntPointer(t, event.NativeVirtualKeyCode, 8)
	assertNilStringPointer(t, event.Text)
	assertNilStringPointer(t, event.UnmodifiedText)
}

func TestKeyboardCreatePressEventPrintableKey(t *testing.T) {
	keyboard := &Keyboard{}

	event := keyboard.createPressEvent("keyDown", "a")

	assertStringPointer(t, event.Code, "KeyA")
	assertStringPointer(t, event.Key, "a")
	assertIntPointer(t, event.WindowsVirtualKeyCode, 65)
	assertIntPointer(t, event.NativeVirtualKeyCode, 65)
	assertStringPointer(t, event.Text, "a")
	assertStringPointer(t, event.UnmodifiedText, "a")
}

func TestKeyboardCreatePressEventUnknownKey(t *testing.T) {
	keyboard := &Keyboard{}

	event := keyboard.createPressEvent("keyDown", "UnknownKey")

	assertNilStringPointer(t, event.Code)
	assertNilStringPointer(t, event.Key)
	assertNilIntPointer(t, event.Modifiers)
	assertNilIntPointer(t, event.WindowsVirtualKeyCode)
	assertNilIntPointer(t, event.NativeVirtualKeyCode)
	assertNilIntPointer(t, event.Location)
	assertNilStringPointer(t, event.Text)
	assertNilStringPointer(t, event.UnmodifiedText)
}

func assertStringPointer(t *testing.T, value *string, expected string) {
	t.Helper()

	if value == nil {
		t.Fatalf("expected %q, got nil", expected)
	}

	if *value != expected {
		t.Fatalf("expected %q, got %q", expected, *value)
	}
}

func assertNilStringPointer(t *testing.T, value *string) {
	t.Helper()

	if value != nil {
		t.Fatalf("expected nil, got %q", *value)
	}
}

func assertIntPointer(t *testing.T, value *int, expected int) {
	t.Helper()

	if value == nil {
		t.Fatalf("expected %d, got nil", expected)
	}

	if *value != expected {
		t.Fatalf("expected %d, got %d", expected, *value)
	}
}

func assertNilIntPointer(t *testing.T, value *int) {
	t.Helper()

	if value != nil {
		t.Fatalf("expected nil, got %d", *value)
	}
}
