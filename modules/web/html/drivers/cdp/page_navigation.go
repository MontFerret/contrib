package cdp

import (
	"context"
	"errors"
	"regexp"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/page"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/dom"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/eval"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/events"
	cdpnet "github.com/MontFerret/contrib/modules/web/html/drivers/cdp/network"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/templates"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func (p *HTMLPage) Navigate(ctx context.Context, url runtime.String) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.network.Navigate(ctx, url); err != nil {
		return err
	}

	return p.loadMainFrame(ctx)
}

func (p *HTMLPage) NavigateBack(ctx context.Context, skip runtime.Int) (runtime.Boolean, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	ret, err := p.network.NavigateBack(ctx, skip)
	if err != nil {
		return runtime.False, err
	}

	return ret, p.loadMainFrame(ctx)
}

func (p *HTMLPage) NavigateForward(ctx context.Context, skip runtime.Int) (runtime.Boolean, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	ret, err := p.network.NavigateForward(ctx, skip)
	if err != nil {
		return runtime.False, err
	}

	return ret, p.loadMainFrame(ctx)
}

func (p *HTMLPage) WaitForNavigation(ctx context.Context, targetURL runtime.String) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	pattern, err := p.urlToRegexp(targetURL)
	if err != nil {
		return err
	}

	return p.waitForNavigation(ctx, cdpnet.WaitEventOptions{URL: pattern})
}

func (p *HTMLPage) WaitForFrameNavigation(ctx context.Context, frame drivers.HTMLDocument, targetURL runtime.String) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	current := p.dom.GetMainFrame()
	doc, ok := frame.(*dom.HTMLDocument)
	if !ok {
		return errors.New("invalid frame type")
	}

	pattern, err := p.urlToRegexp(targetURL)
	if err != nil {
		return err
	}

	frameID := doc.Frame().Frame.ID
	isMain := current.Frame().Frame.ID == frameID
	opts := cdpnet.WaitEventOptions{
		URL: pattern,
	}

	// if it's the current document
	if !isMain {
		opts.FrameID = frameID
	}

	return p.waitForNavigation(ctx, opts)
}

func (p *HTMLPage) urlToRegexp(targetURL runtime.String) (*regexp.Regexp, error) {
	if targetURL == "" {
		return nil, nil
	}

	r, err := regexp.Compile(targetURL.String())
	if err != nil {
		return nil, runtime.Error(err, "invalid URL pattern")
	}

	return r, nil
}

func (p *HTMLPage) navigationStream(ctx context.Context) (runtime.Stream, error) {
	stream, err := p.network.OnNavigation(ctx)
	if err != nil {
		return nil, err
	}

	return newPreparedNavigationEventStream(stream, p.prepareNavigationEvent), nil
}

func (p *HTMLPage) prepareNavigationEvent(ctx context.Context, evt *cdpnet.NavigationEvent) error {
	if evt == nil {
		return nil
	}

	client := evt.SourceClient()
	if client == nil {
		client = p.client
	}

	p.dom.RecordFrameClient(evt.FrameID, client)

	if err := p.waitForDocumentReady(ctx, client, evt.FrameID); err != nil {
		return err
	}

	return p.dom.ReloadRoot(ctx)
}

func (p *HTMLPage) waitForMainFrameReady(ctx context.Context) error {
	ftRepl, err := p.client.Page.GetFrameTree(ctx)
	if err != nil {
		return err
	}

	return p.waitForDocumentReady(ctx, p.client, ftRepl.FrameTree.Frame.ID)
}

func (p *HTMLPage) waitForDocumentReady(ctx context.Context, client *cdp.Client, frameID page.FrameID) error {
	exec, err := eval.Create(ctx, p.logger, client, frameID)
	if err != nil {
		return err
	}

	_, err = events.NewEvalWaitTask(exec, templates.DOMReady(), events.DefaultPolling).Run(ctx)

	return err
}

func (p *HTMLPage) waitForNavigation(ctx context.Context, opts cdpnet.WaitEventOptions) error {
	stream, err := p.navigationStream(ctx)
	if err != nil {
		return err
	}

	defer stream.Close()

	for evt := range stream.Read(ctx) {
		if err := evt.Err(); err != nil {
			return err
		}

		nav, ok := evt.Value().(*cdpnet.NavigationEvent)
		if !ok || !matchNavigationEvent(nav, opts) {
			continue
		}

		return nil
	}

	return ctx.Err()
}

func (p *HTMLPage) parseNavigationSubscriptionOptions(ctx context.Context, value runtime.Map) (cdpnet.WaitEventOptions, error) {
	opts := cdpnet.WaitEventOptions{}
	if value == nil {
		return opts, nil
	}

	frame, err := value.Get(ctx, runtime.NewString("frame"))
	if err == nil && frame != runtime.None {
		frameID, err := navigationFrameID(frame)
		if err != nil {
			return opts, err
		}

		opts.FrameID = frameID
	}

	target, err := value.Get(ctx, runtime.NewString("target"))
	if err == nil && target != runtime.None {
		targetURL, err := runtime.CastString(target)
		if err != nil {
			return opts, err
		}

		pattern, err := p.urlToRegexp(targetURL)
		if err != nil {
			return opts, err
		}

		opts.URL = pattern
	}

	return opts, nil
}
