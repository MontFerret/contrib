package network

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/mafredri/cdp/protocol/network"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/rs/zerolog/log"

	"github.com/MontFerret/ferret/v2/pkg/runtime"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
)

func toDriverBody(body *string) []byte {
	if body == nil {
		return nil
	}

	return []byte(*body)
}

func toDriverBodyEntries(entries []network.PostDataEntry) []byte {
	if len(entries) == 0 {
		return nil
	}

	var body strings.Builder

	for _, entry := range entries {
		if entry.Bytes == nil {
			continue
		}

		body.WriteString(*entry.Bytes)
	}

	if body.Len() == 0 {
		return nil
	}

	return []byte(body.String())
}

func toDriverHeaders(headers network.Headers) *drivers.HTTPHeaders {
	result := drivers.NewHTTPHeaders()
	deserialized := make(map[string]string)

	if len(headers) > 0 {
		err := json.Unmarshal(headers, &deserialized)

		if err != nil {
			log.Trace().Err(err).Msg("failed to deserialize responseReceivedEvent headers")
		}
	}

	ctx := context.Background()

	for key, value := range deserialized {
		_ = result.Set(ctx, runtime.String(key), runtime.String(value))
	}

	return result
}

func toDriverResponse(resp network.Response, body []byte) *drivers.HTTPResponse {
	return &drivers.HTTPResponse{
		URL:          resp.URL,
		StatusCode:   resp.Status,
		Status:       resp.StatusText,
		Headers:      toDriverHeaders(resp.Headers),
		Body:         body,
		ResponseTime: float64(resp.ResponseTime),
	}
}

func toDriverRequest(req network.Request) *drivers.HTTPRequest {
	body := toDriverBodyEntries(req.PostDataEntries)

	if body == nil {
		//lint:ignore SA1019 Older CDP endpoints may omit postDataEntries and only populate PostData.
		body = toDriverBody(req.PostData)
	}

	return &drivers.HTTPRequest{
		URL:     req.URL,
		Method:  req.Method,
		Headers: toDriverHeaders(req.Headers),
		Body:    body,
	}
}

func fromDriverCookie(url string, cookie drivers.HTTPCookie) network.CookieParam {
	sameSite := network.CookieSameSiteNotSet

	switch cookie.SameSite {
	case drivers.SameSiteLaxMode:
		sameSite = network.CookieSameSiteLax
	case drivers.SameSiteStrictMode:
		sameSite = network.CookieSameSiteStrict
	}

	normalizedURL := normalizeCookieURL(url)
	param := network.CookieParam{
		URL:   &normalizedURL,
		Name:  cookie.Name,
		Value: cookie.Value,
	}

	if cookie.Path != "" {
		path := cookie.Path
		param.Path = &path
	}

	if cookie.Domain != "" {
		domain := cookie.Domain
		param.Domain = &domain
	}

	if cookie.Secure {
		secure := true
		param.Secure = &secure
	}

	if cookie.HTTPOnly {
		httpOnly := true
		param.HTTPOnly = &httpOnly
	}

	if sameSite != network.CookieSameSiteNotSet {
		param.SameSite = sameSite
	}

	switch {
	case !cookie.Expires.IsZero():
		param.Expires = network.TimeSinceEpoch(cookie.Expires.Unix())
	case cookie.MaxAge > 0:
		param.Expires = network.TimeSinceEpoch(time.Now().Add(time.Duration(cookie.MaxAge) * time.Second).Unix())
	}

	return param
}

func fromDriverCookieDelete(url string, cookie drivers.HTTPCookie) *network.DeleteCookiesArgs {
	normalizedURL := normalizeCookieURL(url)
	args := network.NewDeleteCookiesArgs(cookie.Name).SetURL(normalizedURL)

	if cookie.Path != "" {
		args.SetPath(cookie.Path)
	}

	if cookie.Domain != "" {
		args.SetDomain(cookie.Domain)
	}

	return args
}

func toDriverCookie(c network.Cookie) drivers.HTTPCookie {
	sameSite := drivers.SameSiteDefaultMode

	switch c.SameSite {
	case network.CookieSameSiteLax:
		sameSite = drivers.SameSiteLaxMode
	case network.CookieSameSiteStrict:
		sameSite = drivers.SameSiteStrictMode
	}

	return drivers.HTTPCookie{
		Name:     c.Name,
		Value:    c.Value,
		Path:     c.Path,
		Domain:   c.Domain,
		Expires:  time.Unix(int64(c.Expires), 0),
		SameSite: sameSite,
		Secure:   c.Secure,
		HTTPOnly: c.HTTPOnly,
	}
}

func normalizeCookieURL(url string) string {
	const httpPrefix = "http://"
	const httpsPrefix = "https://"

	if strings.HasPrefix(url, httpPrefix) || strings.HasPrefix(url, httpsPrefix) {
		return url
	}

	return httpPrefix + url
}

func normalizeCookieLookupURL(url string) (string, bool) {
	if url == "" || url == BlankPageURL {
		return "", false
	}

	return normalizeCookieURL(url), true
}

func isURLMatched(url string, pattern *regexp.Regexp) bool {
	var matched bool

	// if a URL pattern is provided
	if pattern != nil {
		matched = pattern.MatchString(url)
	} else {
		// otherwise, just match
		matched = true
	}

	return matched
}

func isFrameMatched(current, target page.FrameID) bool {
	// if frameID is empty string or equals to the current one
	if len(target) == 0 {
		return true
	}

	return target == current
}
