package dom

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func (doc *HTMLDocument) Dispatch(ctx context.Context, event runtime.DispatchEvent) error {
	return dispatchHTMLDocument(ctx, doc, event)
}
