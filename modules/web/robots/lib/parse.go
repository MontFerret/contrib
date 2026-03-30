package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/robots/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Parse parses raw robots.txt content into a plain object.
func Parse(_ context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 1); err != nil {
		return nil, err
	}

	text, err := runtime.CastArgAt[runtime.String](args, 0)
	if err != nil {
		return nil, err
	}

	doc, err := core.Parse(text.String())
	if err != nil {
		return nil, err
	}

	return encodeValue(doc), nil
}
