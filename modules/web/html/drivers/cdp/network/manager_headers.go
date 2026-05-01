package network

import (
	"context"

	"github.com/goccy/go-json"
	"github.com/mafredri/cdp/protocol/network"
	"github.com/pkg/errors"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
)

func (m *Manager) GetHeaders(_ context.Context) (*drivers.HTTPHeaders, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.headers == nil {
		return drivers.NewHTTPHeaders(), nil
	}

	return m.headers.Clone(), nil
}

func (m *Manager) SetHeaders(ctx context.Context, headers *drivers.HTTPHeaders) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.setHeadersInternal(ctx, headers)
}

func (m *Manager) setHeadersInternal(ctx context.Context, headers *drivers.HTTPHeaders) error {
	m.logger.Trace().Msg("starting to set headers")

	if len(headers.Data) == 0 {
		m.logger.Trace().Msg("no headers passed")

		return nil
	}

	m.headers = headers

	m.logger.Trace().Msg("marshaling headers")

	j, err := json.Marshal(headers)
	if err != nil {
		m.logger.Trace().Err(err).Msg("failed to marshal headers")

		return errors.Wrap(err, "failed to marshal headers")
	}

	m.logger.Trace().Msg("sending headers to browser")

	err = m.client.Network.SetExtraHTTPHeaders(
		ctx,
		network.NewSetExtraHTTPHeadersArgs(j),
	)
	if err != nil {
		m.logger.Trace().Err(err).Msg("failed to set headers")

		return errors.Wrap(err, "failed to set headers")
	}

	m.logger.Trace().Msg("succeeded to set headers")

	return nil
}
