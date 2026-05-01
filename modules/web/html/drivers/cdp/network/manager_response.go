package network

import (
	"context"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/network"

	cdpsession "github.com/MontFerret/contrib/modules/web/html/drivers/cdp/session"
)

func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.Trace().Msg("closing")

	if m.stop != nil {
		m.stop()
		m.stop = nil
	}

	if m.sessions != nil {
		m.sessions.RemoveListener(m.responseListener)
	}

	m.closeResponseWatchers()

	return nil
}

func (m *Manager) handleResponse(msg *network.ResponseReceivedReply) {
	if msg == nil {
		return
	}

	// we are interested in documents only
	if msg.Type != network.ResourceTypeDocument {
		return
	}

	if msg.FrameID == nil {
		return
	}

	log := m.logger.With().
		Str("frame_id", string(*msg.FrameID)).
		Str("request_id", string(msg.RequestID)).
		Str("loader_id", string(msg.LoaderID)).
		Float64("timestamp", float64(msg.Timestamp)).
		Str("url", msg.Response.URL).
		Int("status_code", msg.Response.Status).
		Str("status_text", msg.Response.StatusText).
		Logger()

	log.Trace().Msg("received browser response")

	m.response.Store(*msg.FrameID, toDriverResponse(msg.Response, nil))

	log.Trace().Msg("updated frame response information")
}

func (m *Manager) startResponseWatcher(ctx context.Context) error {
	if m.sessions == nil {
		return m.watchResponseClient(ctx, "root", m.client)
	}

	for _, client := range m.sessions.Snapshot() {
		if err := m.watchResponseStream(ctx, client); err != nil {
			return err
		}
	}

	m.responseListener = m.sessions.AddListener(func(event cdpsession.Event) {
		switch event.Kind {
		case cdpsession.EventAttached:
			if err := m.watchResponseStream(ctx, event.Client); err != nil {
				m.logger.Warn().Err(err).Msg("failed to watch response stream for attached session")
			}
		case cdpsession.EventDetached:
			m.closeResponseStream(event.Client)
		}
	})

	return nil
}

func (m *Manager) watchResponseStream(ctx context.Context, client *cdpsession.Client) error {
	if client == nil || client.CDP == nil {
		return nil
	}

	return m.watchResponseClient(ctx, string(client.ID), client.CDP)
}

func (m *Manager) watchResponseClient(ctx context.Context, key string, client *cdp.Client) error {
	if client == nil {
		return nil
	}

	m.responseMu.Lock()
	if _, exists := m.responseWatchers[key]; exists {
		m.responseMu.Unlock()
		return nil
	}
	m.responseMu.Unlock()

	stream, err := client.Network.ResponseReceived(ctx)
	if err != nil {
		return err
	}

	m.responseMu.Lock()
	m.responseWatchers[key] = stream
	m.responseMu.Unlock()

	go func() {
		defer m.closeResponseWatcher(key)

		for {
			select {
			case <-ctx.Done():
				return
			case <-stream.Ready():
				if ctx.Err() != nil {
					return
				}

				reply, err := stream.Recv()
				if err != nil {
					if ctx.Err() != nil {
						return
					}

					m.logger.Trace().Err(err).Msg("failed to receive response event")

					return
				}

				m.handleResponse(reply)
			}
		}
	}()

	return nil
}

func (m *Manager) closeResponseStream(client *cdpsession.Client) {
	if client == nil {
		return
	}

	m.closeResponseWatcher(string(client.ID))
}

func (m *Manager) closeResponseWatcher(key string) {
	m.responseMu.Lock()
	stream, exists := m.responseWatchers[key]
	if exists {
		delete(m.responseWatchers, key)
	}
	m.responseMu.Unlock()

	if exists {
		_ = stream.Close()
	}
}

func (m *Manager) closeResponseWatchers() {
	m.responseMu.Lock()
	streams := make([]network.ResponseReceivedClient, 0, len(m.responseWatchers))
	for key, stream := range m.responseWatchers {
		streams = append(streams, stream)
		delete(m.responseWatchers, key)
	}
	m.responseMu.Unlock()

	for _, stream := range streams {
		_ = stream.Close()
	}
}
