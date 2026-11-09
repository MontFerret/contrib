package cdp

import (
	"context"
	"errors"
)

func (p *HTMLPage) Close() error {
	if p.policySession != nil {
		return p.policySession.Close()
	}

	return p.closeResources()
}

func (p *HTMLPage) closeResources() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true

	var errs []error
	var url string
	if p.dom != nil {
		frame := p.dom.GetMainFrame()

		if frame != nil {
			url = frame.GetURL().String()
		}
	}

	if p.dom != nil {
		if err := p.dom.Close(); err != nil {
			errs = append(errs, err)
			p.logger.Warn().
				Str("url", url).
				Err(err).
				Msg("failed to close dom manager")
		}
	}

	if p.network != nil {
		if err := p.network.Close(); err != nil {
			errs = append(errs, err)
			p.logger.Warn().
				Str("url", url).
				Err(err).
				Msg("failed to close network manager")
		}
	}

	if p.policySession == nil && p.client != nil && p.client.Page != nil {
		if err := p.client.Page.Close(context.Background()); err != nil {
			p.logger.Warn().
				Str("url", url).
				Err(err).
				Msg("failed to close browser page")
		}
	}

	if p.policySession == nil && p.sessions != nil {
		if err := p.sessions.Close(); err != nil {
			p.logger.Warn().
				Str("url", url).
				Err(err).
				Msg("failed to close session manager")
		}
	}

	if p.policySession != nil {
		return errors.Join(errs...)
	}

	return nil
}

func (p *HTMLPage) withPolicyContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if p.policySession == nil {
		return ctx, func() {}
	}

	operationCtx, cancel := context.WithCancel(ctx)
	stop := context.AfterFunc(p.policySession.lifetime, cancel)

	return operationCtx, func() {
		stop()
		cancel()
	}
}
