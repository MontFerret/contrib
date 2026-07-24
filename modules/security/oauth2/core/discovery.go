package core

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	ferrethttp "github.com/MontFerret/ferret/v2/pkg/net/http"
)

// DiscoveryOptions controls authorization-server metadata discovery.
type DiscoveryOptions struct {
	InsecureAllowHTTP bool
	Timeout           time.Duration
}

// Discover loads and validates RFC 8414 authorization-server metadata.
func Discover(
	ctx context.Context,
	httpClient ferrethttp.Client,
	issuer string,
	options DiscoveryOptions,
) (*Provider, error) {
	if httpClient == nil {
		return nil, fmt.Errorf("oauth2 discovery HTTP client is required")
	}

	if options.Timeout < 0 {
		return nil, fmt.Errorf("oauth2 discovery timeout must not be negative")
	}

	discoveryURL, err := wellKnownURL(issuer, options.InsecureAllowHTTP)
	if err != nil {
		return nil, fmt.Errorf("oauth2 discovery issuer is invalid: %w", err)
	}

	if ctx == nil {
		ctx = context.Background()
	}

	requestCtx := ctx
	cancel := func() {}

	if options.Timeout > 0 {
		requestCtx, cancel = context.WithTimeout(ctx, options.Timeout)
	}

	defer cancel()

	response, err := httpClient.Do(requestCtx, &ferrethttp.Request{
		Method: http.MethodGet,
		URL:    discoveryURL,
		Headers: ferrethttp.Headers{
			"Accept": {"application/json"},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("oauth2 discovery request: %w", err)
	}

	if response == nil {
		return nil, fmt.Errorf("oauth2 discovery request returned no response")
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"oauth2 discovery request returned HTTP status %d",
			response.StatusCode,
		)
	}

	if err := validateJSONContentType(response.Headers); err != nil {
		return nil, fmt.Errorf("oauth2 discovery response: %w", err)
	}

	config, err := decodeDiscoveryMetadata(response.Body, issuer, options.InsecureAllowHTTP)
	if err != nil {
		return nil, fmt.Errorf("oauth2 discovery response: %w", err)
	}

	provider, err := NewProvider(config)
	if err != nil {
		return nil, fmt.Errorf("oauth2 discovery response: %w", err)
	}

	return provider, nil
}

func decodeDiscoveryMetadata(
	body []byte,
	expectedIssuer string,
	insecureAllowHTTP bool,
) (ProviderConfig, error) {
	var fields map[string]json.RawMessage
	decoder := json.NewDecoder(bytes.NewReader(body))

	if err := decoder.Decode(&fields); err != nil {
		return ProviderConfig{}, fmt.Errorf("metadata is not valid JSON: %w", err)
	}

	if fields == nil {
		return ProviderConfig{}, fmt.Errorf("metadata must be a JSON object")
	}

	var trailing any

	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return ProviderConfig{}, fmt.Errorf("metadata must contain one JSON object")
		}

		return ProviderConfig{}, fmt.Errorf("metadata has invalid trailing content: %w", err)
	}

	issuer, present, err := decodeMetadataString(fields, "issuer")
	if err != nil {
		return ProviderConfig{}, err
	}

	if !present || issuer == "" {
		return ProviderConfig{}, fmt.Errorf("metadata issuer is required")
	}

	if issuer != expectedIssuer {
		return ProviderConfig{}, fmt.Errorf(
			"metadata issuer %q does not exactly match expected issuer %q",
			issuer,
			expectedIssuer,
		)
	}

	tokenEndpoint, present, err := decodeMetadataString(fields, "token_endpoint")
	if err != nil {
		return ProviderConfig{}, err
	}

	if !present || tokenEndpoint == "" {
		return ProviderConfig{}, fmt.Errorf("metadata token_endpoint is required")
	}

	authorizationEndpoint, _, err := decodeMetadataString(fields, "authorization_endpoint")
	if err != nil {
		return ProviderConfig{}, err
	}

	revocationEndpoint, _, err := decodeMetadataString(fields, "revocation_endpoint")
	if err != nil {
		return ProviderConfig{}, err
	}

	introspectionEndpoint, _, err := decodeMetadataString(fields, "introspection_endpoint")
	if err != nil {
		return ProviderConfig{}, err
	}

	jwksURI, _, err := decodeMetadataString(fields, "jwks_uri")
	if err != nil {
		return ProviderConfig{}, err
	}

	scopes, _, err := decodeMetadataStrings(fields, "scopes_supported")
	if err != nil {
		return ProviderConfig{}, err
	}

	grants, grantsPresent, err := decodeMetadataStrings(fields, "grant_types_supported")
	if err != nil {
		return ProviderConfig{}, err
	}

	if !grantsPresent {
		grants = []string{"authorization_code", "implicit"}
	}

	authMethods, authMethodsPresent, err := decodeMetadataStrings(
		fields,
		"token_endpoint_auth_methods_supported",
	)
	if err != nil {
		return ProviderConfig{}, err
	}

	if !authMethodsPresent {
		authMethods = []string{string(ClientAuthMethodBasic)}
	}

	return ProviderConfig{
		Issuer:                   issuer,
		AuthorizationEndpoint:    authorizationEndpoint,
		TokenEndpoint:            tokenEndpoint,
		RevocationEndpoint:       revocationEndpoint,
		IntrospectionEndpoint:    introspectionEndpoint,
		JWKSURI:                  jwksURI,
		ScopesSupported:          scopes,
		GrantTypesSupported:      grants,
		TokenEndpointAuthMethods: authMethods,
		InsecureAllowHTTP:        insecureAllowHTTP,
	}, nil
}

func decodeMetadataString(
	fields map[string]json.RawMessage,
	name string,
) (string, bool, error) {
	raw, present := fields[name]
	if !present {
		return "", false, nil
	}

	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return "", true, fmt.Errorf("metadata %s must be a string", name)
	}

	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", true, fmt.Errorf("metadata %s must be a string: %w", name, err)
	}

	return value, true, nil
}

func decodeMetadataStrings(
	fields map[string]json.RawMessage,
	name string,
) ([]string, bool, error) {
	raw, present := fields[name]
	if !present {
		return nil, false, nil
	}

	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil, true, fmt.Errorf("metadata %s must be an array of strings", name)
	}

	var values []string
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, true, fmt.Errorf("metadata %s must be an array of strings: %w", name, err)
	}

	return values, true, nil
}

func validateJSONContentType(headers ferrethttp.Headers) error {
	contentType := firstHeaderValue(headers, "Content-Type")
	if contentType == "" {
		return fmt.Errorf("Content-Type must be application/json")
	}

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return fmt.Errorf("Content-Type is invalid: %w", err)
	}

	if !strings.EqualFold(mediaType, "application/json") {
		return fmt.Errorf("Content-Type must be application/json, got %q", mediaType)
	}

	return nil
}

func firstHeaderValue(headers ferrethttp.Headers, name string) string {
	for key, values := range headers {
		if strings.EqualFold(key, name) && len(values) > 0 {
			return values[0]
		}
	}

	return ""
}
