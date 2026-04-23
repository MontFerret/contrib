package network

import (
	"context"

	"github.com/mafredri/cdp/protocol/network"
	"github.com/pkg/errors"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

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
