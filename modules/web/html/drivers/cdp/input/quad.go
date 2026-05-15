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

type Quad struct {
	X float64
	Y float64
}

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

func GetClickablePointByObjectID(ctx context.Context, client *cdp.Client, objectID runtime.RemoteObjectID) (Quad, error) {
	return getClickablePoint(ctx, client, dom.NewGetContentQuadsArgs().SetObjectID(objectID))
}

func GetElementPointByObjectID(ctx context.Context, client *cdp.Client, objectID runtime.RemoteObjectID, xOffset, yOffset *float64) (Quad, error) {
	return getElementPoint(ctx, client, dom.NewGetContentQuadsArgs().SetObjectID(objectID), xOffset, yOffset)
}
