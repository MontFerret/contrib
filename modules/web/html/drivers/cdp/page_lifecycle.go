package cdp

import "context"

func (p *HTMLPage) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var url string
	frame := p.dom.GetMainFrame()

	if frame != nil {
		url = frame.GetURL().String()
	}

	p.closed = true

	err := p.dom.Close()
	if err != nil {
		p.logger.Warn().
			Str("url", url).
			Err(err).
			Msg("failed to close dom manager")
	}

	err = p.network.Close()
	if err != nil {
		p.logger.Warn().
			Str("url", url).
			Err(err).
			Msg("failed to close network manager")
	}

	err = p.client.Page.Close(context.Background())
	if err != nil {
		p.logger.Warn().
			Str("url", url).
			Err(err).
			Msg("failed to close browser page")
	}

	if err := p.sessions.Close(); err != nil {
		p.logger.Warn().
			Str("url", url).
			Err(err).
			Msg("failed to close session manager")
	}

	return nil
}
