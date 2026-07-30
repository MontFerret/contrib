package input

import (
	"context"
	"math"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/dom"
	"github.com/mafredri/cdp/protocol/runtime"
	"github.com/pkg/errors"

	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/utils"
)

type (
	Quad struct {
		X float64
		Y float64
	}

	elementBounds struct {
		left   float64
		top    float64
		right  float64
		bottom float64
	}

	mouseMoveCandidate struct {
		point    Quad
		distance float64
		viable   bool
	}
)

func fromProtocolQuad(quad dom.Quad) []Quad {
	return []Quad{
		{
			X: quad[0],
			Y: quad[1],
		},
		{
			X: quad[2],
			Y: quad[3],
		},
		{
			X: quad[4],
			Y: quad[5],
		},
		{
			X: quad[6],
			Y: quad[7],
		},
	}
}

func newMouseMoveCandidate(cursor, point Quad, viewportWidth, viewportHeight float64) mouseMoveCandidate {
	xDelta := point.X - cursor.X
	yDelta := point.Y - cursor.Y

	return mouseMoveCandidate{
		point:    point,
		distance: xDelta*xDelta + yDelta*yDelta,
		viable:   point.X >= 0 && point.X <= viewportWidth && point.Y >= 0 && point.Y <= viewportHeight,
	}
}

func computeQuadArea(quads []Quad) float64 {
	var area float64

	for i := range quads {
		p1 := quads[i]
		p2 := quads[(i+1)%len(quads)]
		area += (p1.X*p2.Y - p2.X*p1.Y) / 2
	}

	return math.Abs(area)
}

func intersectQuadWithViewport(quad []Quad, width, height float64) []Quad {
	quads := make([]Quad, 0, len(quad))

	for _, point := range quad {
		quads = append(quads, Quad{
			X: math.Min(math.Max(point.X, 0), width),
			Y: math.Min(math.Max(point.Y, 0), height),
		})
	}

	return quads
}

func getClickablePoint(ctx context.Context, client *cdp.Client, qargs *dom.GetContentQuadsArgs) (Quad, error) {
	contentQuadsReply, err := client.DOM.GetContentQuads(ctx, qargs)

	if err != nil {
		return Quad{}, err
	}

	if len(contentQuadsReply.Quads) == 0 {
		return Quad{}, errors.New("node is either not visible or not an HTMLElement")
	}

	layoutMetricsReply, err := client.Page.GetLayoutMetrics(ctx)

	if err != nil {
		return Quad{}, err
	}

	clientWidth, clientHeight := utils.GetLayoutViewportWH(layoutMetricsReply)
	quads := make([][]Quad, 0, len(contentQuadsReply.Quads))

	for _, q := range contentQuadsReply.Quads {
		quad := intersectQuadWithViewport(fromProtocolQuad(q), float64(clientWidth), float64(clientHeight))

		if computeQuadArea(quad) > 1 {
			quads = append(quads, quad)
		}
	}

	if len(quads) == 0 {
		return Quad{}, errors.New("node is either not visible or not an HTMLElement")
	}

	// Return the middle point of the first quad.
	quad := quads[0]
	var x float64
	var y float64

	for _, q := range quad {
		x += q.X
		y += q.Y
	}

	return Quad{
		X: x / 4,
		Y: y / 4,
	}, nil
}

func getElementPoint(ctx context.Context, client *cdp.Client, qargs *dom.GetContentQuadsArgs, xOffset, yOffset *float64) (Quad, error) {
	contentQuadsReply, err := client.DOM.GetContentQuads(ctx, qargs)

	if err != nil {
		return Quad{}, err
	}

	if len(contentQuadsReply.Quads) == 0 {
		return Quad{}, errors.New("node is either not visible or not an HTMLElement")
	}

	layoutMetricsReply, err := client.Page.GetLayoutMetrics(ctx)

	if err != nil {
		return Quad{}, err
	}

	clientWidth, clientHeight := utils.GetLayoutViewportWH(layoutMetricsReply)

	for _, protocolQuad := range contentQuadsReply.Quads {
		quad := intersectQuadWithViewport(fromProtocolQuad(protocolQuad), float64(clientWidth), float64(clientHeight))

		if computeQuadArea(quad) <= 1 {
			continue
		}

		var centerX, centerY float64
		left := quad[0].X
		top := quad[0].Y

		for _, point := range quad {
			centerX += point.X
			centerY += point.Y
			left = math.Min(left, point.X)
			top = math.Min(top, point.Y)
		}

		x := centerX / 4
		y := centerY / 4

		if xOffset != nil {
			x = left + *xOffset
		}

		if yOffset != nil {
			y = top + *yOffset
		}

		return Quad{X: x, Y: y}, nil
	}

	return Quad{}, errors.New("node is either not visible or not an HTMLElement")
}

func getElementBounds(ctx context.Context, client *cdp.Client, qargs *dom.GetContentQuadsArgs) (elementBounds, float64, float64, error) {
	contentQuadsReply, err := client.DOM.GetContentQuads(ctx, qargs)
	if err != nil {
		return elementBounds{}, 0, 0, err
	}

	if len(contentQuadsReply.Quads) == 0 {
		return elementBounds{}, 0, 0, errors.New("node is either not visible or not an HTMLElement")
	}

	layoutMetricsReply, err := client.Page.GetLayoutMetrics(ctx)
	if err != nil {
		return elementBounds{}, 0, 0, err
	}

	clientWidth, clientHeight := utils.GetLayoutViewportWH(layoutMetricsReply)
	var bounds elementBounds
	found := false

	for _, protocolQuad := range contentQuadsReply.Quads {
		quad := intersectQuadWithViewport(fromProtocolQuad(protocolQuad), float64(clientWidth), float64(clientHeight))

		if computeQuadArea(quad) <= 1 {
			continue
		}

		for _, point := range quad {
			if !found {
				bounds = elementBounds{
					left:   point.X,
					top:    point.Y,
					right:  point.X,
					bottom: point.Y,
				}
				found = true
				continue
			}

			bounds.left = math.Min(bounds.left, point.X)
			bounds.top = math.Min(bounds.top, point.Y)
			bounds.right = math.Max(bounds.right, point.X)
			bounds.bottom = math.Max(bounds.bottom, point.Y)
		}
	}

	if !found {
		return elementBounds{}, 0, 0, errors.New("node is either not visible or not an HTMLElement")
	}

	return bounds, float64(clientWidth), float64(clientHeight), nil
}

func getMouseMovePoint(bounds elementBounds, cursor Quad, viewportWidth, viewportHeight, distance float64) Quad {
	horizontalY := math.Min(math.Max(cursor.Y, bounds.top), bounds.bottom)
	verticalX := math.Min(math.Max(cursor.X, bounds.left), bounds.right)
	candidates := []mouseMoveCandidate{
		newMouseMoveCandidate(cursor, Quad{X: bounds.left - distance, Y: horizontalY}, viewportWidth, viewportHeight),
		newMouseMoveCandidate(cursor, Quad{X: bounds.right + distance, Y: horizontalY}, viewportWidth, viewportHeight),
		newMouseMoveCandidate(cursor, Quad{X: verticalX, Y: bounds.top - distance}, viewportWidth, viewportHeight),
		newMouseMoveCandidate(cursor, Quad{X: verticalX, Y: bounds.bottom + distance}, viewportWidth, viewportHeight),
	}

	best := candidates[0]

	for _, candidate := range candidates[1:] {
		switch {
		case candidate.viable && !best.viable:
			best = candidate
		case candidate.viable == best.viable && candidate.distance < best.distance:
			best = candidate
		}
	}

	return best.point
}

func getMouseMovePointByObjectID(
	ctx context.Context,
	client *cdp.Client,
	objectID runtime.RemoteObjectID,
	cursor Quad,
	distance float64,
) (Quad, error) {
	bounds, viewportWidth, viewportHeight, err := getElementBounds(
		ctx,
		client,
		dom.NewGetContentQuadsArgs().SetObjectID(objectID),
	)
	if err != nil {
		return Quad{}, err
	}

	return getMouseMovePoint(bounds, cursor, viewportWidth, viewportHeight, distance), nil
}

func getClickablePointByObjectID(ctx context.Context, client *cdp.Client, objectID runtime.RemoteObjectID) (Quad, error) {
	return getClickablePoint(ctx, client, dom.NewGetContentQuadsArgs().SetObjectID(objectID))
}

func getElementPointByObjectID(ctx context.Context, client *cdp.Client, objectID runtime.RemoteObjectID, xOffset, yOffset *float64) (Quad, error) {
	return getElementPoint(ctx, client, dom.NewGetContentQuadsArgs().SetObjectID(objectID), xOffset, yOffset)
}
