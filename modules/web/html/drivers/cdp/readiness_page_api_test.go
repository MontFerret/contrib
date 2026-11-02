package cdp

import (
	"context"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/page"
)

type readinessPageAPI struct {
	cdp.Page
	createWorld func(context.Context, *page.CreateIsolatedWorldArgs) (*page.CreateIsolatedWorldReply, error)
	frameTree   page.FrameTree
}

func (api *readinessPageAPI) CreateIsolatedWorld(ctx context.Context, args *page.CreateIsolatedWorldArgs) (*page.CreateIsolatedWorldReply, error) {
	return api.createWorld(ctx, args)
}

func (api *readinessPageAPI) GetFrameTree(context.Context) (*page.GetFrameTreeReply, error) {
	return &page.GetFrameTreeReply{FrameTree: api.frameTree}, nil
}
