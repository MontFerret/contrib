package cdp

import (
	"context"
	"errors"
)

// Close uses the standalone cleanup budget. With managed cleanup the driver
// already owns the page and performs browser cleanup when the host closes it.
func (p *HTMLPage) Close() error {
	if p.policySession != nil && p.policySession.managed() {
		p.policySession.invalidate()

		return nil
	}

	ctx, cancel := defaultCleanupContext()
	defer cancel()

	return p.CloseContext(ctx)
}

// CloseContext starts or waits for page cleanup. The first caller supplies
// the cleanup context; each caller can independently cancel its wait.
// Once cleanup settles, repeated calls return its result.
func (p *HTMLPage) CloseContext(ctx context.Context) error {
	if p.policySession != nil {
		return p.policySession.CloseContext(ctx)
	}

	return p.cleanup.Run(ctx, p.close)
}

func (p *HTMLPage) close(ctx context.Context) error {
	var err error
	if p.sessions != nil {
		err = p.sessions.ClosePageTargets(ctx, nil)
		managerErr := p.sessions.CloseContext(ctx)
		if ctx.Err() != nil {
			managerErr = p.sessions.CloseContext(context.Background())
		}

		err = errors.Join(err, managerErr)
	} else if p.client != nil && p.client.Page != nil {
		if err = ctx.Err(); err == nil {
			err = p.client.Page.Close(ctx)
		}
	}

	return errors.Join(err, p.closeResources())
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

	return errors.Join(errs...)
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
