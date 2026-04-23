package network

import (
	"context"

	"github.com/mafredri/cdp/protocol/page"
	"github.com/pkg/errors"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func (m *Manager) GetResponse(_ context.Context, frameID page.FrameID) (drivers.HTTPResponse, error) {
	value, found := m.response.Load(frameID)

	m.logger.Trace().
		Str("frame_id", string(frameID)).
		Bool("found", found).
		Msg("getting frame response")

	if !found {
		return drivers.HTTPResponse{}, runtime.ErrNotFound
	}

	return *(value.(*drivers.HTTPResponse)), nil
}

func (m *Manager) Navigate(ctx context.Context, url runtime.String) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if url == "" {
		url = BlankPageURL
	}

	urlStr := url.String()
	m.logger.Trace().Str("url", urlStr).Msg("starting navigation")

	repl, err := m.client.Page.Navigate(ctx, page.NewNavigateArgs(urlStr))
	if err == nil && repl.ErrorText != nil {
		err = errors.New(*repl.ErrorText)
	}

	if err != nil {
		m.logger.Trace().Err(err).Msg("failed starting navigation")

		return err
	}

	m.logger.Trace().Msg("succeeded starting navigation")

	return m.WaitForNavigation(ctx, WaitEventOptions{})
}

func (m *Manager) NavigateForward(ctx context.Context, skip runtime.Int) (runtime.Boolean, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.Trace().
		Int64("skip", int64(skip)).
		Msg("starting forward navigation")

	history, err := m.client.Page.GetNavigationHistory(ctx)
	if err != nil {
		m.logger.Trace().
			Err(err).
			Msg("failed to get navigation history")

		return runtime.False, err
	}

	length := len(history.Entries)
	lastIndex := length - 1

	// nowhere to go forward
	if history.CurrentIndex == lastIndex {
		m.logger.Trace().
			Int("history_entries", length).
			Int("history_current_index", history.CurrentIndex).
			Int("history_last_index", lastIndex).
			Msg("no forward history. nowhere to navigate. done.")

		return runtime.False, nil
	}

	if skip < 1 {
		skip = 1
	}

	to := int(skip) + history.CurrentIndex
	if to > lastIndex {
		m.logger.Trace().
			Int("history_entries", length).
			Int("history_current_index", history.CurrentIndex).
			Int("history_last_index", lastIndex).
			Int("history_target_index", to).
			Msg("not enough history items. using the edge index")

		to = lastIndex
	}

	entry := history.Entries[to]
	err = m.client.Page.NavigateToHistoryEntry(ctx, page.NewNavigateToHistoryEntryArgs(entry.ID))
	if err != nil {
		m.logger.Trace().
			Int("history_entries", length).
			Int("history_current_index", history.CurrentIndex).
			Int("history_last_index", lastIndex).
			Int("history_target_index", to).
			Err(err).
			Msg("failed to get navigation history entry")

		return runtime.False, err
	}

	err = m.WaitForNavigation(ctx, WaitEventOptions{})
	if err != nil {
		m.logger.Trace().
			Int("history_entries", length).
			Int("history_current_index", history.CurrentIndex).
			Int("history_last_index", lastIndex).
			Int("history_target_index", to).
			Err(err).
			Msg("failed to wait for navigation completion")

		return runtime.False, err
	}

	m.logger.Trace().
		Int("history_entries", length).
		Int("history_current_index", history.CurrentIndex).
		Int("history_last_index", lastIndex).
		Int("history_target_index", to).
		Msg("succeeded to wait for navigation completion")

	return runtime.True, nil
}

func (m *Manager) NavigateBack(ctx context.Context, skip runtime.Int) (runtime.Boolean, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.Trace().
		Int64("skip", int64(skip)).
		Msg("starting backward navigation")

	history, err := m.client.Page.GetNavigationHistory(ctx)
	if err != nil {
		m.logger.Trace().Err(err).Msg("failed to get navigation history")

		return runtime.False, err
	}

	length := len(history.Entries)

	// we are in the beginning
	if history.CurrentIndex == 0 {
		m.logger.Trace().
			Int("history_entries", length).
			Int("history_current_index", history.CurrentIndex).
			Msg("no backward history. nowhere to navigate. done.")

		return runtime.False, nil
	}

	if skip < 1 {
		skip = 1
	}

	to := history.CurrentIndex - int(skip)
	if to < 0 {
		m.logger.Trace().
			Int("history_entries", length).
			Int("history_current_index", history.CurrentIndex).
			Int("history_target_index", to).
			Msg("not enough history items. using 0 index")

		to = 0
	}

	entry := history.Entries[to]
	err = m.client.Page.NavigateToHistoryEntry(ctx, page.NewNavigateToHistoryEntryArgs(entry.ID))
	if err != nil {
		m.logger.Trace().
			Int("history_entries", length).
			Int("history_current_index", history.CurrentIndex).
			Int("history_target_index", to).
			Err(err).
			Msg("failed to get navigation history entry")

		return runtime.False, err
	}

	err = m.WaitForNavigation(ctx, WaitEventOptions{})
	if err != nil {
		m.logger.Trace().
			Int("history_entries", length).
			Int("history_current_index", history.CurrentIndex).
			Int("history_target_index", to).
			Err(err).
			Msg("failed to wait for navigation completion")

		return runtime.False, err
	}

	m.logger.Trace().
		Int("history_entries", length).
		Int("history_current_index", history.CurrentIndex).
		Int("history_target_index", to).
		Msg("succeeded to wait for navigation completion")

	return runtime.True, nil
}

func (m *Manager) WaitForNavigation(ctx context.Context, opts WaitEventOptions) error {
	stream, err := m.OnNavigation(ctx)
	if err != nil {
		return err
	}

	defer stream.Close()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for evt := range stream.Read(ctx) {
		if err := evt.Err(); err != nil {
			return err
		}

		nav := evt.Value().(*NavigationEvent)
		if !isFrameMatched(nav.FrameID, opts.FrameID) || !isURLMatched(nav.URL, opts.URL) {
			continue
		}

		return nil
	}

	return ctx.Err()
}

func (m *Manager) OnNavigation(ctx context.Context) (runtime.Stream, error) {
	return newNavigationEventStream(m.logger, m.sessions), nil
}

func (m *Manager) OnRequest(ctx context.Context) (runtime.Stream, error) {
	return newRequestEventStream(m.logger, m.sessions), nil
}

func (m *Manager) OnResponse(ctx context.Context) (runtime.Stream, error) {
	return newResponseEventStream(m.logger, m.sessions), nil
}
