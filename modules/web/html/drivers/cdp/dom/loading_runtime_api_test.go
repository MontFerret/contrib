package dom

import (
	"context"

	"github.com/mafredri/cdp"
	cdpruntime "github.com/mafredri/cdp/protocol/runtime"
)

type loadingRuntimeAPI struct {
	cdp.Runtime
	call func(context.Context, *cdpruntime.CallFunctionOnArgs) (*cdpruntime.CallFunctionOnReply, error)
}

func (api *loadingRuntimeAPI) CallFunctionOn(ctx context.Context, args *cdpruntime.CallFunctionOnArgs) (*cdpruntime.CallFunctionOnReply, error) {
	return api.call(ctx, args)
}
