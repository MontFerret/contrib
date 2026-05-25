package cdp

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func (p *HTMLPage) Query(ctx context.Context, q runtime.Query) (runtime.List, error) {
	return p.getCurrentDocument().Query(ctx, q)
}

func (p *HTMLPage) QueryOne(ctx context.Context, q runtime.Query) (runtime.Value, error) {
	return p.getCurrentDocument().QueryOne(ctx, q)
}

func (p *HTMLPage) QueryCount(ctx context.Context, q runtime.Query) (runtime.Int, error) {
	return p.getCurrentDocument().QueryCount(ctx, q)
}

func (p *HTMLPage) QueryExists(ctx context.Context, q runtime.Query) (runtime.Boolean, error) {
	return p.getCurrentDocument().QueryExists(ctx, q)
}
