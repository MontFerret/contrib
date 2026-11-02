package dom

import (
	"context"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/page"
)

type loadingPageAPI struct {
	cdp.Page
	frameTree   func(context.Context) (*page.GetFrameTreeReply, error)
	createWorld func(context.Context, *page.CreateIsolatedWorldArgs) (*page.CreateIsolatedWorldReply, error)
}

func (api *loadingPageAPI) GetFrameTree(ctx context.Context) (*page.GetFrameTreeReply, error) {
	return api.frameTree(ctx)
}

func (api *loadingPageAPI) CreateIsolatedWorld(ctx context.Context, args *page.CreateIsolatedWorldArgs) (*page.CreateIsolatedWorldReply, error) {
	return api.createWorld(ctx, args)
}
