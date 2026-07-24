package core

import (
	"fmt"
	"net/url"
	"strings"
)

var sensitiveTokenEndpointQueryKeys = map[string]struct{}{
	"access_token":          {},
	"assertion":             {},
	"authorization":         {},
	"client_assertion":      {},
	"client_assertion_type": {},
	"client_id":             {},
	"client_secret":         {},
	"code":                  {},
	"code_verifier":         {},
	"id_token":              {},
	"refresh_token":         {},
}

func parseProviderURL(field string, rawURL string, insecureAllowHTTP bool) (*url.URL, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("oauth2 provider %s is invalid: %w", field, err)
	}

	if parsed.Opaque != "" || !parsed.IsAbs() || parsed.Host == "" || parsed.Hostname() == "" {
		return nil, fmt.Errorf("oauth2 provider %s must be an absolute HTTP(S) URL", field)
	}

	if parsed.User != nil {
		return nil, fmt.Errorf("oauth2 provider %s must not contain user information", field)
	}

	if strings.Contains(rawURL, "#") {
		return nil, fmt.Errorf("oauth2 provider %s must not contain a fragment", field)
	}

	switch {
	case strings.EqualFold(parsed.Scheme, "https"):
	case strings.EqualFold(parsed.Scheme, "http") && insecureAllowHTTP:
	case strings.EqualFold(parsed.Scheme, "http"):
		return nil, fmt.Errorf("oauth2 provider %s must use HTTPS", field)
	default:
		return nil, fmt.Errorf("oauth2 provider %s must use HTTP or HTTPS", field)
	}

	if _, err := url.ParseQuery(parsed.RawQuery); err != nil {
		return nil, fmt.Errorf("oauth2 provider %s has an invalid query: %w", field, err)
	}

	return parsed, nil
}

func parseIssuerURL(rawURL string, insecureAllowHTTP bool) (*url.URL, error) {
	parsed, err := parseProviderURL("issuer", rawURL, insecureAllowHTTP)
	if err != nil {
		return nil, err
	}

	if parsed.RawQuery != "" || parsed.ForceQuery {
		return nil, fmt.Errorf("oauth2 provider issuer must not contain a query")
	}

	return parsed, nil
}

func parseTokenEndpointURL(rawURL string, insecureAllowHTTP bool) (*url.URL, error) {
	parsed, err := parseProviderURL("token endpoint", rawURL, insecureAllowHTTP)
	if err != nil {
		return nil, err
	}

	query, err := url.ParseQuery(parsed.RawQuery)
	if err != nil {
		return nil, fmt.Errorf("oauth2 provider token endpoint has an invalid query: %w", err)
	}

	for key := range query {
		normalized := strings.ToLower(key)
		_, sensitive := sensitiveTokenEndpointQueryKeys[normalized]

		if sensitive || strings.HasPrefix(normalized, "client_assertion") {
			return nil, fmt.Errorf(
				"oauth2 provider token endpoint must not contain sensitive query parameter %q",
				key,
			)
		}
	}

	return parsed, nil
}

func wellKnownURL(issuer string, insecureAllowHTTP bool) (string, error) {
	parsed, err := parseIssuerURL(issuer, insecureAllowHTTP)
	if err != nil {
		return "", err
	}

	const wellKnownPath = "/.well-known/oauth-authorization-server"
	escapedIssuerPath := parsed.EscapedPath()
	parsed.Path = wellKnownPath + parsed.Path
	parsed.RawPath = wellKnownPath + escapedIssuerPath

	return parsed.String(), nil
}
