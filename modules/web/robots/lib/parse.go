package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/robots/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// Parse parses raw robots.txt content into a plain object.
func Parse(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 1); err != nil {
		return nil, err
	}

	text, err := sdk.DecodeArg[runtime.String](ctx, args, 0)
	if err != nil {
		return nil, err
	}

	doc, err := core.Parse(text.String())
	if err != nil {
		return nil, err
	}

	return sdk.Encode(ctx, doc)
}
