package core

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	stdhttp "net/http"
	"net/url"
	"time"

	ferrethttp "github.com/MontFerret/ferret/v2/pkg/net/http"
)

// Executor performs OAuth token grants with an injected Ferret HTTP client.
type Executor struct {
	httpClient ferrethttp.Client
	clock      Clock
}

// NewExecutor constructs a token grant executor.
func NewExecutor(httpClient ferrethttp.Client, options ...ExecutorOption) (*Executor, error) {
	if tokenDependencyIsNil(httpClient) {
		return nil, fmt.Errorf("%w: HTTP client is required", ErrInvalidExecutor)
	}

	executor := &Executor{
		httpClient: httpClient,
		clock:      ClockFunc(time.Now),
	}

	for _, option := range options {
		if option == nil {
			continue
		}

		if err := option(executor); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrInvalidExecutor, err)
		}
	}

	return executor, nil
}

// ClientCredentials exchanges client credentials for a token.
func (e *Executor) ClientCredentials(
	ctx context.Context,
	client *Client,
	options ClientCredentialsOptions,
) (*TokenSet, error) {
	snapshot, err := e.validateClient(client)
	if err != nil {
		return nil, err
	}

	if snapshot.AuthMethod == ClientAuthMethodNone {
		return nil, fmt.Errorf(
			"%w: client_credentials requires an authenticated client",
			ErrInvalidGrant,
		)
	}

	if options.Timeout < 0 {
		return nil, fmt.Errorf("%w: timeout must not be negative", ErrInvalidGrant)
	}

	parameters := cloneTokenParameters(options.Parameters)
	reserved := map[string]struct{}{
		"grant_type": {},
		"scope":      {},
		"audience":   {},
	}

	if err := validateTokenParameters(parameters, reserved); err != nil {
		return nil, err
	}

	scope, err := encodeTokenScope(append([]string(nil), options.Scope...))
	if err != nil {
		return nil, err
	}

	parameters["grant_type"] = []string{"client_credentials"}
	if scope != "" {
		parameters["scope"] = []string{scope}
	}

	if options.Audience != "" {
		parameters["audience"] = []string{options.Audience}
	}

	return e.executeTokenRequest(ctx, snapshot, parameters, options.Timeout, scope, "")
}

// Refresh exchanges a refresh token for a new token set.
func (e *Executor) Refresh(
	ctx context.Context,
	client *Client,
	previous *TokenSet,
	options RefreshOptions,
) (*TokenSet, error) {
	snapshot, err := e.validateClient(client)
	if err != nil {
		return nil, err
	}

	if options.Timeout < 0 {
		return nil, fmt.Errorf("%w: timeout must not be negative", ErrInvalidGrant)
	}

	previous = previous.Clone()
	refreshToken := options.RefreshToken

	if refreshToken == "" && previous != nil {
		refreshToken = previous.RefreshToken
	}

	if refreshToken == "" {
		return nil, fmt.Errorf("%w: refresh token is required", ErrInvalidGrant)
	}

	parameters := cloneTokenParameters(options.Parameters)
	reserved := map[string]struct{}{
		"grant_type":    {},
		"refresh_token": {},
		"scope":         {},
	}

	if err := validateTokenParameters(parameters, reserved); err != nil {
		return nil, err
	}

	scope, err := encodeTokenScope(append([]string(nil), options.Scope...))
	if err != nil {
		return nil, err
	}

	scopeFallback := scope
	if previous != nil && previous.Scope != "" {
		if err := validateTokenScopeString(previous.Scope); err != nil {
			return nil, fmt.Errorf("%w: previous token scope is malformed", ErrInvalidGrant)
		}

		if !tokenScopesAreSubset(options.Scope, previous.Scope) {
			return nil, fmt.Errorf(
				"%w: refresh scope must not include scopes absent from the previous token",
				ErrInvalidGrant,
			)
		}

		if scopeFallback == "" {
			scopeFallback = previous.Scope
		}
	}

	parameters["grant_type"] = []string{"refresh_token"}
	parameters["refresh_token"] = []string{refreshToken}

	if scope != "" {
		parameters["scope"] = []string{scope}
	}

	return e.executeTokenRequest(
		ctx,
		snapshot,
		parameters,
		options.Timeout,
		scopeFallback,
		refreshToken,
	)
}

// Token performs an OAuth extension grant. grant_type must appear exactly once
// in Parameters and must be an absolute URI.
func (e *Executor) Token(
	ctx context.Context,
	client *Client,
	options TokenOptions,
) (*TokenSet, error) {
	snapshot, err := e.validateClient(client)
	if err != nil {
		return nil, err
	}

	if options.Timeout < 0 {
		return nil, fmt.Errorf("%w: timeout must not be negative", ErrInvalidGrant)
	}

	parameters := cloneTokenParameters(options.Parameters)
	grantTypes, found := parameters["grant_type"]

	if !found || len(grantTypes) != 1 {
		return nil, fmt.Errorf(
			"%w: grant_type must appear exactly once",
			ErrInvalidGrant,
		)
	}

	if err := validateExtensionGrantType(grantTypes[0]); err != nil {
		return nil, err
	}

	delete(parameters, "grant_type")

	if err := validateTokenParameters(parameters, nil); err != nil {
		return nil, err
	}

	parameters["grant_type"] = append([]string(nil), grantTypes...)

	scopeFallback, err := tokenScopeFallback(parameters)
	if err != nil {
		return nil, err
	}

	return e.executeTokenRequest(
		ctx,
		snapshot,
		parameters,
		options.Timeout,
		scopeFallback,
		"",
	)
}

func (e *Executor) validateClient(client *Client) (*Client, error) {
	if e == nil || tokenDependencyIsNil(e.httpClient) || tokenDependencyIsNil(e.clock) {
		return nil, fmt.Errorf("%w: executor is not initialized", ErrInvalidExecutor)
	}

	snapshot := client.Clone()

	if err := snapshot.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidGrant, err)
	}

	return snapshot, nil
}

func (e *Executor) executeTokenRequest(
	ctx context.Context,
	client *Client,
	parameters Parameters,
	timeout time.Duration,
	scopeFallback string,
	refreshTokenFallback string,
) (*TokenSet, error) {
	headers := ferrethttp.Headers{
		"Accept":       {"application/json"},
		"Content-Type": {"application/x-www-form-urlencoded"},
	}

	if err := e.applyClientAuthentication(client, parameters, headers); err != nil {
		return nil, err
	}

	form := make(url.Values, len(parameters))

	for key, values := range parameters {
		form[key] = append([]string(nil), values...)
	}

	requestCtx := ctx

	if requestCtx == nil {
		requestCtx = context.Background()
	}

	cancel := func() {}

	if timeout > 0 {
		requestCtx, cancel = context.WithTimeout(requestCtx, timeout)
	}

	defer cancel()

	secrets := tokenSubmittedSecrets(client, parameters, headers)
	response, err := e.httpClient.Do(requestCtx, &ferrethttp.Request{
		Method:  stdhttp.MethodPost,
		URL:     client.Provider.TokenEndpoint,
		Headers: headers,
		Body:    []byte(form.Encode()),
	})

	if err != nil {
		return nil, newTokenRequestError(err, secrets)
	}

	token, err := parseTokenEndpointResponse(
		response,
		e.clock.Now(),
		scopeFallback,
		refreshTokenFallback,
		secrets,
	)

	if err != nil {
		if oauthError, ok := errors.AsType[*Error](err); ok {
			return nil, oauthError
		}

		return nil, newTokenRequestError(err, secrets)
	}

	return token, nil
}

func (e *Executor) applyClientAuthentication(
	client *Client,
	parameters Parameters,
	headers ferrethttp.Headers,
) error {
	switch client.AuthMethod {
	case ClientAuthMethodBasic:
		encodedID := url.QueryEscape(client.ClientID)
		encodedSecret := url.QueryEscape(client.ClientSecret)
		credentials := base64.StdEncoding.EncodeToString(
			[]byte(encodedID + ":" + encodedSecret),
		)
		headers["Authorization"] = []string{"Basic " + credentials}
	case ClientAuthMethodPost:
		headers["Authorization"] = nil
		parameters["client_id"] = []string{client.ClientID}
		parameters["client_secret"] = []string{client.ClientSecret}
	case ClientAuthMethodNone:
		headers["Authorization"] = nil
		parameters["client_id"] = []string{client.ClientID}
	default:
		return fmt.Errorf("%w: unsupported client authentication method", ErrInvalidGrant)
	}

	return nil
}
