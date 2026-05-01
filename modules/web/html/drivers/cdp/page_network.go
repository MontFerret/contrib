package cdp

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
)

func (p *HTMLPage) GetCookies(ctx context.Context) (*drivers.HTTPCookies, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.network.GetCookies(ctx, p.getCurrentDocument().GetURL().String())
}

func (p *HTMLPage) SetCookies(ctx context.Context, cookies *drivers.HTTPCookies) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.network.SetCookies(ctx, p.getCurrentDocument().GetURL().String(), cookies)
}

func (p *HTMLPage) DeleteCookies(ctx context.Context, cookies *drivers.HTTPCookies) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.network.DeleteCookies(ctx, p.getCurrentDocument().GetURL().String(), cookies)
}

func (p *HTMLPage) GetResponse(ctx context.Context) (drivers.HTTPResponse, error) {
	doc := p.getCurrentDocument()
	if doc == nil {
		return drivers.HTTPResponse{}, nil
	}

	return p.network.GetResponse(ctx, doc.Frame().Frame.ID)
}
