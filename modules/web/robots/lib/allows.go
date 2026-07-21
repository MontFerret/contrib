package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/robots/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// Allows returns whether the path is allowed for the given user-agent.
func Allows(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 2, 3); err != nil {
		return nil, err
	}

	doc, err := sdk.DecodeArg[core.Document](ctx, args, 0)
	if err != nil {
		return nil, err
	}

	path, err := sdk.DecodeArg[runtime.String](ctx, args, 1)
	if err != nil {
		return nil, err
	}

	userAgent := "*"
	if len(args) > 2 {
		raw, err := sdk.DecodeArg[runtime.String](ctx, args, 2)
		if err != nil {
			return nil, err
		}

		userAgent = raw.String()
	}

	return runtime.NewBoolean(core.Allows(doc, path.String(), userAgent)), nil
}
