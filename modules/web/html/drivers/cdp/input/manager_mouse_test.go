package input

import (
	"context"
	"errors"
	"testing"

	"github.com/mafredri/cdp"
	"github.com/rs/zerolog"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestMoveMouseByXYMovesWithinViewport(t *testing.T) {
	t.Parallel()

	pageAPI := newMouseLayoutPage(800, 600, nil)
	inputAPI := newMouseDispatchInput(nil)
	manager, mouse := newMouseMoveManager(pageAPI, inputAPI)

	moved, err := manager.MoveMouseByXY(context.Background(), 125.5, 250.25)
	if err != nil {
		t.Fatalf("move mouse: %v", err)
	}

	if moved != runtime.True {
		t.Fatalf("expected mouse to move, got %s", moved)
	}

	if pageAPI.calls != 1 {
		t.Fatalf("expected one layout metrics call, got %d", pageAPI.calls)
	}

	assertLastMouseMove(t, inputAPI, 125.5, 250.25)

	if mouse.x != 125.5 || mouse.y != 250.25 {
		t.Fatalf("expected tracked cursor at (125.5, 250.25), got (%v, %v)", mouse.x, mouse.y)
	}
}

func TestMoveMouseByXYClampsToViewportPixels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		startX   float64
		startY   float64
		inputX   runtime.Float
		inputY   runtime.Float
		expected Quad
	}{
		{
			name:     "negative coordinates",
			startX:   50,
			startY:   40,
			inputX:   -10,
			inputY:   -20,
			expected: Quad{X: 0, Y: 0},
		},
		{
			name:     "coordinates beyond viewport",
			inputX:   1000,
			inputY:   1000,
			expected: Quad{X: 99, Y: 79},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			pageAPI := newMouseLayoutPage(100, 80, nil)
			inputAPI := newMouseDispatchInput(nil)
			manager, mouse := newMouseMoveManager(pageAPI, inputAPI)
			mouse.x = tc.startX
			mouse.y = tc.startY

			moved, err := manager.MoveMouseByXY(context.Background(), tc.inputX, tc.inputY)
			if err != nil {
				t.Fatalf("move mouse: %v", err)
			}

			if moved != runtime.True {
				t.Fatalf("expected mouse to move, got %s", moved)
			}

			assertLastMouseMove(t, inputAPI, tc.expected.X, tc.expected.Y)
		})
	}
}

func TestMoveMouseByXYSkipsNoOp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		width  int
		height int
		start  Quad
		input  Quad
	}{
		{
			name:   "repeated coordinates",
			width:  100,
			height: 80,
			start:  Quad{X: 25, Y: 30},
			input:  Quad{X: 25, Y: 30},
		},
		{
			name:   "already at clamped border",
			width:  100,
			height: 80,
			start:  Quad{X: 99, Y: 79},
			input:  Quad{X: 1000, Y: 1000},
		},
		{
			name:  "empty viewport",
			start: Quad{X: 0, Y: 0},
			input: Quad{X: 10, Y: 10},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			pageAPI := newMouseLayoutPage(tc.width, tc.height, nil)
			inputAPI := newMouseDispatchInput(nil)
			manager, mouse := newMouseMoveManager(pageAPI, inputAPI)
			mouse.x = tc.start.X
			mouse.y = tc.start.Y

			moved, err := manager.MoveMouseByXY(
				context.Background(),
				runtime.Float(tc.input.X),
				runtime.Float(tc.input.Y),
			)
			if err != nil {
				t.Fatalf("move mouse: %v", err)
			}

			if moved != runtime.False {
				t.Fatalf("expected no-op result, got %s", moved)
			}

			if len(inputAPI.calls) != 0 {
				t.Fatalf("expected no mouse dispatch, got %d calls", len(inputAPI.calls))
			}
		})
	}
}

func TestMoveMouseByXYReturnsMetricsError(t *testing.T) {
	t.Parallel()

	metricsErr := errors.New("metrics failed")
	pageAPI := newMouseLayoutPage(100, 80, metricsErr)
	inputAPI := newMouseDispatchInput(nil)
	manager, _ := newMouseMoveManager(pageAPI, inputAPI)

	moved, err := manager.MoveMouseByXY(context.Background(), 10, 20)
	if !errors.Is(err, metricsErr) {
		t.Fatalf("expected metrics error, got %v", err)
	}

	if moved != runtime.False {
		t.Fatalf("expected false on metrics failure, got %s", moved)
	}

	if len(inputAPI.calls) != 0 {
		t.Fatalf("expected no mouse dispatch, got %d calls", len(inputAPI.calls))
	}
}

func TestMoveMouseByXYReturnsDispatchError(t *testing.T) {
	t.Parallel()

	dispatchErr := errors.New("dispatch failed")
	pageAPI := newMouseLayoutPage(100, 80, nil)
	inputAPI := newMouseDispatchInput(dispatchErr)
	manager, mouse := newMouseMoveManager(pageAPI, inputAPI)

	moved, err := manager.MoveMouseByXY(context.Background(), 10, 20)
	if !errors.Is(err, dispatchErr) {
		t.Fatalf("expected dispatch error, got %v", err)
	}

	if moved != runtime.False {
		t.Fatalf("expected false on dispatch failure, got %s", moved)
	}

	if len(inputAPI.calls) == 0 {
		t.Fatal("expected a mouse dispatch attempt")
	}

	if mouse.x != 0 || mouse.y != 0 {
		t.Fatalf("expected tracked cursor to remain at (0, 0), got (%v, %v)", mouse.x, mouse.y)
	}
}

func newMouseMoveManager(pageAPI cdp.Page, inputAPI cdp.Input) (*Manager, *Mouse) {
	client := &cdp.Client{
		Page:  pageAPI,
		Input: inputAPI,
	}
	mouse := NewMouse(client)

	return New(zerolog.Nop(), client, nil, nil, mouse), mouse
}

func assertLastMouseMove(t *testing.T, inputAPI *mouseDispatchInput, expectedX, expectedY float64) {
	t.Helper()

	if len(inputAPI.calls) == 0 {
		t.Fatal("expected a mouse dispatch")
	}

	last := inputAPI.calls[len(inputAPI.calls)-1]

	if last.Type != "mouseMoved" {
		t.Fatalf("expected mouseMoved event, got %q", last.Type)
	}

	if last.X != expectedX || last.Y != expectedY {
		t.Fatalf("expected mouse move to (%v, %v), got (%v, %v)", expectedX, expectedY, last.X, last.Y)
	}
}
