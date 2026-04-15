package network

import (
	"context"
	"sync"

	"github.com/goccy/go-json"
	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/network"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	cdpsession "github.com/MontFerret/contrib/modules/web/html/drivers/cdp/session"
	"github.com/MontFerret/contrib/modules/web/html/drivers/common"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

const BlankPageURL = "about:blank"

type (
	FrameLoadedListener = func(ctx context.Context, frame page.Frame)

	Manager struct {
		logger           zerolog.Logger
		client           *cdp.Client
		sessions         *cdpsession.Manager
		headers          *drivers.HTTPHeaders
		interceptor      *Interceptor
		stop             context.CancelFunc
		response         *sync.Map
		responseWatchers map[string]network.ResponseReceivedClient
		responseListener cdpsession.ListenerID
		mu               sync.RWMutex
		responseMu       sync.Mutex
	}
)

func New(
	logger zerolog.Logger,
	client *cdp.Client,
	sessions *cdpsession.Manager,
	options Options,
) (*Manager, error) {
	ctx, cancel := context.WithCancel(context.Background())

	m := new(Manager)
	m.logger = common.LoggerWithName(logger.With(), "network_manager").Logger()
	m.client = client
	m.sessions = sessions
	m.headers = drivers.NewHTTPHeaders()
	m.stop = cancel
	m.response = new(sync.Map)
	m.responseWatchers = make(map[string]network.ResponseReceivedClient)

	var err error

	defer func() {
		if err != nil {
			m.stop()
		}
	}()

	if options.Filter != nil && len(options.Filter.Patterns) > 0 {
		m.interceptor = NewInterceptor(logger, client)

		if err := m.interceptor.AddFilter("resources", options.Filter); err != nil {
			return nil, err
		}

		if err = m.interceptor.Run(ctx); err != nil {
			return nil, err
		}
	}

	if len(options.Cookies) > 0 {
		for url, cookies := range options.Cookies {
			err = m.setCookiesInternal(ctx, url, cookies)

			if err != nil {
				return nil, err
			}
		}
	}

	if options.Headers != nil && len(options.Headers.Data) > 0 {
		err = m.setHeadersInternal(ctx, options.Headers)

		if err != nil {
			return nil, err
		}
	}

	if err = m.startResponseWatcher(ctx); err != nil {
		return nil, err
	}

	return m, nil
}

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

func (m *Manager) GetCookies(ctx context.Context, url string) (*drivers.HTTPCookies, error) {
	m.logger.Trace().Str("url", url).Msg("starting to get cookies")

	args := network.NewGetCookiesArgs()

	if normalizedURL, ok := normalizeCookieLookupURL(url); ok {
		args.SetURLs([]string{normalizedURL})
	}

	repl, err := m.client.Network.GetCookies(ctx, args)

	if err != nil {
		m.logger.Trace().Err(err).Msg("failed to get cookies")

		return nil, errors.Wrap(err, "failed to get cookies")
	}

	cookies := drivers.NewHTTPCookies()

	if repl.Cookies == nil {
		m.logger.Trace().Msg("no cookies found")

		return cookies, nil
	}

	for _, c := range repl.Cookies {
		cookie := toDriverCookie(c)
		_ = cookies.Set(ctx, runtime.String(cookie.Name), cookie)
	}

	m.logger.Trace().Msg("succeeded to get cookies")

	return cookies, nil
}

func (m *Manager) SetCookies(ctx context.Context, url string, cookies *drivers.HTTPCookies) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.setCookiesInternal(ctx, url, cookies)
}

func (m *Manager) setCookiesInternal(ctx context.Context, url string, cookies *drivers.HTTPCookies) error {
	m.logger.Trace().Str("url", url).Msg("starting to set cookies")

	if cookies == nil {
		m.logger.Trace().Msg("nil cookies passed")

		return errors.Wrap(runtime.ErrMissedArgument, "cookies")
	}

	if len(cookies.Data) == 0 {
		m.logger.Trace().Msg("no cookies passed")

		return nil
	}

	params := make([]network.CookieParam, 0, len(cookies.Data))

	for name, cookie := range cookies.Data {
		m.logger.Trace().Str("name", name).Msg("preparing a cookie")

		params = append(params, fromDriverCookie(url, cookie))
	}

	err := m.client.Network.SetCookies(ctx, network.NewSetCookiesArgs(params))

	if err != nil {
		m.logger.Trace().Err(err).Msg("failed to set cookies")

		return err
	}

	m.logger.Trace().Msg("succeeded to set cookies")

	return nil
}

func (m *Manager) DeleteCookies(ctx context.Context, url string, cookies *drivers.HTTPCookies) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.Trace().Str("url", url).Msg("starting to delete cookies")

	if cookies == nil {
		m.logger.Trace().Msg("nil cookies passed")

		return errors.Wrap(runtime.ErrMissedArgument, "cookies")
	}

	if len(cookies.Data) == 0 {
		m.logger.Trace().Msg("no cookies passed")

		return nil
	}

	for name, cookie := range cookies.Data {
		m.logger.Trace().Str("name", name).Msg("preparing a cookie for deletion")

		if err := m.client.Network.DeleteCookies(ctx, fromDriverCookieDelete(url, cookie)); err != nil {
			m.logger.Trace().Err(err).Str("name", cookie.Name).Msg("failed to delete a cookie")

			return err
		}

		m.logger.Trace().Str("name", cookie.Name).Msg("succeeded to delete a cookie")
	}

	return nil
}

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
