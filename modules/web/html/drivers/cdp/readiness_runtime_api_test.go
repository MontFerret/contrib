package cdp

import (
	"context"

	"github.com/mafredri/cdp"
	cdpruntime "github.com/mafredri/cdp/protocol/runtime"
)

type readinessRuntimeAPI struct {
	cdp.Runtime
	call     func(context.Context, *cdpruntime.CallFunctionOnArgs) (*cdpruntime.CallFunctionOnReply, error)
	evaluate func(context.Context, *cdpruntime.EvaluateArgs) (*cdpruntime.EvaluateReply, error)
}

func (api *readinessRuntimeAPI) CallFunctionOn(ctx context.Context, args *cdpruntime.CallFunctionOnArgs) (*cdpruntime.CallFunctionOnReply, error) {
	return api.call(ctx, args)
}

func (api *readinessRuntimeAPI) Evaluate(ctx context.Context, args *cdpruntime.EvaluateArgs) (*cdpruntime.EvaluateReply, error) {
	return api.evaluate(ctx, args)
}
