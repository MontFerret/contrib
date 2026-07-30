package input

import (
	"context"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/page"
)

type mouseLayoutPage struct {
	cdp.Page
	reply *page.GetLayoutMetricsReply
	err   error
	calls int
}

func newMouseLayoutPage(width, height int, err error) *mouseLayoutPage {
	return &mouseLayoutPage{
		reply: &page.GetLayoutMetricsReply{
			CSSLayoutViewport: page.LayoutViewport{
				ClientWidth:  width,
				ClientHeight: height,
			},
		},
		err: err,
	}
}

func (p *mouseLayoutPage) GetLayoutMetrics(context.Context) (*page.GetLayoutMetricsReply, error) {
	p.calls++

	return p.reply, p.err
}
