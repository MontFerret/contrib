package input

import (
	"context"
	"sync"
	"time"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/input"
)

type Mouse struct {
	client *cdp.Client
	mu     sync.Mutex
	x      float64
	y      float64
}

func NewMouse(client *cdp.Client) *Mouse {
	return &Mouse{client: client}
}

func (m *Mouse) Click(ctx context.Context, x, y float64, delay time.Duration) error {
	return m.ClickWithCount(ctx, x, y, delay, 1)
}

func (m *Mouse) ClickWithCount(ctx context.Context, x, y float64, delay time.Duration, count int) error {
	return m.ClickWithButton(ctx, x, y, delay, input.MouseButtonLeft, count)
}

func (m *Mouse) ClickWithButton(ctx context.Context, x, y float64, delay time.Duration, button input.MouseButton, count int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.move(ctx, x, y, 1); err != nil {
		return err
	}

	if err := m.button(ctx, "mousePressed", button, count); err != nil {
		return err
	}

	time.Sleep(randomDuration(int(delay)))

	return m.button(ctx, "mouseReleased", button, count)
}

func (m *Mouse) Down(ctx context.Context, button input.MouseButton) error {
	return m.DownWithCount(ctx, button, 1)
}

func (m *Mouse) DownWithCount(ctx context.Context, button input.MouseButton, count int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.button(ctx, "mousePressed", button, count)
}

func (m *Mouse) Up(ctx context.Context, button input.MouseButton) error {
	return m.UpWithCount(ctx, button, 1)
}

func (m *Mouse) UpWithCount(ctx context.Context, button input.MouseButton, count int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.button(ctx, "mouseReleased", button, count)
}

func (m *Mouse) button(ctx context.Context, event string, button input.MouseButton, count int) error {
	return m.client.Input.DispatchMouseEvent(
		ctx,
		input.NewDispatchMouseEventArgs(event, m.x, m.y).
			SetButton(button).
			SetClickCount(count),
	)
}

func (m *Mouse) Move(ctx context.Context, x, y float64) error {
	return m.MoveBySteps(ctx, x, y, 1)
}

func (m *Mouse) MoveBySteps(ctx context.Context, x, y float64, steps int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.move(ctx, x, y, steps)
}

func (m *Mouse) buttonAt(ctx context.Context, event string, point Quad, button input.MouseButton, count int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.move(ctx, point.X, point.Y, 1); err != nil {
		return err
	}

	return m.button(ctx, event, button, count)
}

func (m *Mouse) position() Quad {
	m.mu.Lock()
	defer m.mu.Unlock()

	return Quad{X: m.x, Y: m.y}
}

func (m *Mouse) move(ctx context.Context, x, y float64, steps int) error {
	fromX := m.x
	fromY := m.y

	for i := 0; i <= steps; i++ {
		iFloat := float64(i)
		stepFloat := float64(steps)
		toX := fromX + (x-fromX)*(iFloat/stepFloat)
		toY := fromY + (y-fromY)*(iFloat/stepFloat)

		err := m.client.Input.DispatchMouseEvent(
			ctx,
			input.NewDispatchMouseEventArgs("mouseMoved", toX, toY),
		)

		if err != nil {
			return err
		}
	}

	m.x = x
	m.y = y

	return nil
}
