package cdp

import (
	"context"
	"hash/fnv"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/templates"
	"github.com/MontFerret/contrib/modules/web/html/drivers/internal/data"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func (p *HTMLPage) MarshalJSON() ([]byte, error) {
	return p.getCurrentDocument().MarshalJSON()
}

func (p *HTMLPage) Type() runtime.Type {
	return drivers.HTMLPageType
}

func (p *HTMLPage) String() string {
	return p.getCurrentDocument().GetURL().String()
}

func (p *HTMLPage) Equal(ctx context.Context, other runtime.Value) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

	otherPage, ok := other.(*HTMLPage)
	if !ok {
		return false, nil
	}

	comparison, err := p.compare(ctx, otherPage)

	return comparison == runtime.Equal, err
}

func (p *HTMLPage) Compare(ctx context.Context, other runtime.Value) (runtime.Ordering, error) {
	if err := ctx.Err(); err != nil {
		return runtime.Equal, err
	}

	otherPage, ok := other.(*HTMLPage)
	if ok {
		return p.compare(ctx, otherPage)
	}

	if _, ok := other.(drivers.HTMLPage); ok {
		if err := ctx.Err(); err != nil {
			return runtime.Equal, err
		}

		return runtime.Greater, nil
	}

	return runtime.Equal, runtime.Errorf(
		runtime.ErrInvalidOperation,
		"cannot compare %s with %s",
		runtime.TypeName(runtime.TypeOf(p)),
		runtime.TypeName(runtime.TypeOf(other)),
	)
}

func (p *HTMLPage) compare(ctx context.Context, other *HTMLPage) (runtime.Ordering, error) {
	if err := ctx.Err(); err != nil {
		return runtime.Equal, err
	}

	return runtime.CompareValues(ctx, p.getCurrentDocument().GetURL(), other.getCurrentDocument().GetURL())
}

func (p *HTMLPage) Unwrap() any {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p
}

func (p *HTMLPage) Hash() uint64 {
	h := fnv.New64a()

	h.Write([]byte("CDP"))
	h.Write([]byte(p.Type().String()))
	h.Write([]byte(":"))
	h.Write([]byte(p.getCurrentDocument().GetURL()))

	return h.Sum64()
}

func (p *HTMLPage) Copy() runtime.Value {
	return runtime.None
}

func (p *HTMLPage) Get(ctx context.Context, key runtime.Value) (runtime.Value, error) {
	return data.GetInPage(ctx, key, p)
}

func (p *HTMLPage) Iterate(ctx context.Context) (runtime.Iterator, error) {
	return p.getCurrentDocument().Iterate(ctx)
}

func (p *HTMLPage) Length(ctx context.Context) (runtime.Int, error) {
	return p.getCurrentDocument().Length(ctx)
}

func (p *HTMLPage) IsClosed() runtime.Boolean {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.closed
}

func (p *HTMLPage) GetURL() runtime.String {
	res, err := p.getCurrentDocument().Eval().EvalValue(context.Background(), templates.GetURL())
	if err == nil {
		return runtime.ToString(res)
	}

	p.logger.Warn().
		Err(err).
		Msg("failed to retrieve URL")

	return p.getCurrentDocument().GetURL()
}
