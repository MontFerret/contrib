package cdp

import (
	"context"
)

func (p *HTMLPage) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true

	var url string
	if p.dom != nil {
		frame := p.dom.GetMainFrame()

		if frame != nil {
			url = frame.GetURL().String()
		}
	}

	if p.dom != nil {
		if err := p.dom.Close(); err != nil {
			p.logger.Warn().
				Str("url", url).
				Err(err).
				Msg("failed to close dom manager")
		}
	}

	if p.network != nil {
		if err := p.network.Close(); err != nil {
			p.logger.Warn().
				Str("url", url).
				Err(err).
				Msg("failed to close network manager")
		}
	}

	if p.client != nil && p.client.Page != nil {
		if err := p.client.Page.Close(context.Background()); err != nil {
			p.logger.Warn().
				Str("url", url).
				Err(err).
				Msg("failed to close browser page")
		}
	}

	if p.sessions != nil {
		if err := p.sessions.Close(); err != nil {
			p.logger.Warn().
				Str("url", url).
				Err(err).
				Msg("failed to close session manager")
		}
	}

	return nil
}
