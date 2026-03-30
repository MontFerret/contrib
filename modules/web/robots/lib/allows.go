package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/robots/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Allows returns whether the path is allowed for the given user-agent.
func Allows(_ context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 2, 3); err != nil {
		return nil, err
	}

	doc, err := decodeDocument(args[0])
	if err != nil {
		return nil, err
	}

	path, err := runtime.CastArgAt[runtime.String](args, 1)
	if err != nil {
		return nil, err
	}

	userAgent := "*"
	if len(args) > 2 {
		raw, err := runtime.CastArgAt[runtime.String](args, 2)
		if err != nil {
			return nil, err
		}

		userAgent = raw.String()
	}

	return runtime.NewBoolean(core.Allows(doc, path.String(), userAgent)), nil
}
