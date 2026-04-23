package cdp

import (
	"context"
	"time"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/dom"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func (p *HTMLPage) GetMainFrame() drivers.HTMLDocument {
	return p.getCurrentDocument()
}

func (p *HTMLPage) GetFrames(ctx context.Context) (runtime.List, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.loadFrames(ctx)
}

func (p *HTMLPage) GetFrame(ctx context.Context, idx runtime.Int) (runtime.Value, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	frames, err := p.loadFrames(ctx)
	if err != nil {
		return runtime.None, err
	}

	return frames.At(ctx, idx)
}

func (p *HTMLPage) loadMainFrame(ctx context.Context) error {
	if err := p.waitForMainFrameReady(ctx); err != nil {
		return err
	}

	return p.dom.ReloadRoot(ctx)
}

func (p *HTMLPage) getCurrentDocument() *dom.HTMLDocument {
	return p.dom.GetMainFrame()
}

func (p *HTMLPage) loadFrames(ctx context.Context) (runtime.List, error) {
	expectedFrames := 1

	if doc := p.getCurrentDocument(); doc != nil {
		iframeCount, err := doc.CountBySelector(ctx, drivers.NewCSSSelector("iframe"))
		if err == nil && iframeCount >= 0 {
			expectedFrames += int(iframeCount)
		}
	}

	for attempt := 0; attempt < maxFrameRefreshAttempts; attempt++ {
		if err := p.dom.ReloadRoot(ctx); err != nil {
			return nil, err
		}

		frames, err := p.dom.GetFrameNodes(ctx)
		if err != nil {
			return nil, err
		}

		count, err := frames.Length(ctx)
		if err != nil {
			return nil, err
		}

		if int(count) >= expectedFrames || attempt == maxFrameRefreshAttempts-1 {
			return frames, nil
		}

		timer := time.NewTimer(frameRefreshInterval)

		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}

			return nil, ctx.Err()
		case <-timer.C:
		}
	}

	return p.dom.GetFrameNodes(ctx)
}
