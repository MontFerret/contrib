package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/xml/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// DecodeStream decodes XML content lazily into normalized XML events.
// @param {String|Binary} data - XML content.
// @return {Iterator<Object>} - Proxy exposing an iterator over XML events.
func DecodeStream(_ context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 1); err != nil {
		return nil, err
	}

	content, err := core.ResolveContent(args[0])
	if err != nil {
		return nil, err
	}

	iter, err := core.NewDecodeIterator(content)
	if err != nil {
		return nil, err
	}

	return sdk.NewProxy(iter), nil
}
