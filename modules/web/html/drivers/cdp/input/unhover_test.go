package input

import "testing"

func TestGetUnhoverPointUsesNearestViewportViableEdge(t *testing.T) {
	t.Parallel()

	bounds := elementBounds{
		left:   80,
		top:    80,
		right:  120,
		bottom: 120,
	}
	cases := []struct {
		name   string
		cursor Quad
		want   Quad
	}{
		{
			name:   "left",
			cursor: Quad{X: 81, Y: 100},
			want:   Quad{X: 70, Y: 100},
		},
		{
			name:   "right",
			cursor: Quad{X: 119, Y: 100},
			want:   Quad{X: 130, Y: 100},
		},
		{
			name:   "top",
			cursor: Quad{X: 100, Y: 81},
			want:   Quad{X: 100, Y: 70},
		},
		{
			name:   "bottom",
			cursor: Quad{X: 100, Y: 119},
			want:   Quad{X: 100, Y: 130},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := getMouseMovePoint(bounds, tc.cursor, 200, 200, 10)

			if got != tc.want {
				t.Fatalf("expected point %+v, got %+v", tc.want, got)
			}
		})
	}
}

func TestGetUnhoverPointAvoidsViewportEdge(t *testing.T) {
	t.Parallel()

	bounds := elementBounds{
		left:   2,
		top:    80,
		right:  42,
		bottom: 120,
	}
	cursor := Quad{X: 3, Y: 100}

	got := getMouseMovePoint(bounds, cursor, 200, 200, 10)
	want := Quad{X: 3, Y: 70}

	if got != want {
		t.Fatalf("expected nearest viewport-safe point %+v, got %+v", want, got)
	}
}

func TestGetUnhoverPointFallsBackOutsideViewport(t *testing.T) {
	t.Parallel()

	bounds := elementBounds{
		left:   0,
		top:    0,
		right:  200,
		bottom: 200,
	}
	cursor := Quad{X: 25, Y: 100}

	got := getMouseMovePoint(bounds, cursor, 200, 200, 10)
	want := Quad{X: -10, Y: 100}

	if got != want {
		t.Fatalf("expected nearest outside-viewport point %+v, got %+v", want, got)
	}
}

func TestGetUnhoverPointAlwaysLeavesElementBounds(t *testing.T) {
	t.Parallel()

	bounds := elementBounds{
		left:   40,
		top:    50,
		right:  160,
		bottom: 150,
	}
	cursors := []Quad{
		{X: 41, Y: 51},
		{X: 159, Y: 51},
		{X: 41, Y: 149},
		{X: 159, Y: 149},
		{X: 100, Y: 100},
	}

	for _, cursor := range cursors {
		point := getMouseMovePoint(bounds, cursor, 200, 200, minRndMouseDistance)

		if point.X >= bounds.left &&
			point.X <= bounds.right &&
			point.Y >= bounds.top &&
			point.Y <= bounds.bottom {
			t.Fatalf("expected point %+v to be outside bounds %+v", point, bounds)
		}
	}
}

func TestRandomUnhoverDistance(t *testing.T) {
	t.Parallel()

	var first float64
	var varied bool

	for i := 0; i < 64; i++ {
		distance := randomMouseDistance()

		if distance < minRndMouseDistance || distance >= maxRndMouseDistance {
			t.Fatalf(
				"expected distance in [%v, %v), got %v",
				minRndMouseDistance,
				maxRndMouseDistance,
				distance,
			)
		}

		if i == 0 {
			first = distance
			continue
		}

		if distance != first {
			varied = true
		}
	}

	if !varied {
		t.Fatal("expected unhover distance to vary between samples")
	}
}
