package cdp

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func (p *HTMLPage) Query(ctx context.Context, q runtime.Query) (runtime.List, error) {
	return p.getCurrentDocument().Query(ctx, q)
}
