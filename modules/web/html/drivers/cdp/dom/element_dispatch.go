package dom

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func (el *HTMLElement) Dispatch(ctx context.Context, event runtime.DispatchEvent) error {
	return dispatchHTMLElement(ctx, el, event, false)
}
