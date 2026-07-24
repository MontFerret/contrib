package core

import (
	"context"
	"encoding/base64"
	"errors"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	ferrethttp "github.com/MontFerret/ferret/v2/pkg/net/http"
)

func TestNewExecutorRejectsNilDependencies(t *testing.T) {
	var typedNilClient *recordingTokenHTTPClient

	if _, err := NewExecutor(typedNilClient); !errors.Is(err, ErrInvalidExecutor) {
		t.Fatalf("expected typed nil HTTP client rejection, got %v", err)
	}

	httpClient := newRecordingTokenHTTPClient(
		func(context.Context, *ferrethttp.Request) (*ferrethttp.Response, error) {
			return nil, nil
		},
	)
	var nilClock ClockFunc
	if _, err := NewExecutor(httpClient, WithClock(nilClock)); !errors.Is(
		err,
		ErrInvalidExecutor,
	) {
		t.Fatalf("expected typed nil clock rejection, got %v", err)
	}
}

func TestExecutorClientCredentialsBasic(t *testing.T) {
	var captured *ferrethttp.Request
	client := newRecordingTokenHTTPClient(
		func(_ context.Context, request *ferrethttp.Request) (*ferrethttp.Response, error) {
			captured = request

			return successfulTokenResponse(`{
				"access_token":"access",
				"token_type":"Bearer",
				"expires_in":60,
				"provider_number":1.25
			}`), nil
		},
	)
	now := time.Date(2026, 7, 24, 12, 0, 0, 0, time.UTC)
	executor, err := NewExecutor(client, WithClock(ClockFunc(func() time.Time { return now })))
	if err != nil {
		t.Fatalf("construct executor: %v", err)
	}

	token, err := executor.ClientCredentials(context.Background(), testOAuthClient(
		ClientAuthMethodBasic,
		"client id",
		"s:e/c ret",
	), ClientCredentialsOptions{
		Scope:    []string{"users:read", "projects:read"},
		Audience: "https://api.example.com",
		Parameters: Parameters{
			"resource": {"one", "two"},
		},
	})
	if err != nil {
		t.Fatalf("client credentials: %v", err)
	}

	if captured.Method != "POST" || captured.URL != "https://auth.example.com/token" {
		t.Fatalf("unexpected request target: %#v", captured)
	}
	if got := captured.Headers["Accept"]; !reflect.DeepEqual(got, []string{"application/json"}) {
		t.Fatalf("unexpected Accept header: %v", got)
	}
	if got := captured.Headers["Content-Type"]; !reflect.DeepEqual(
		got,
		[]string{"application/x-www-form-urlencoded"},
	) {
		t.Fatalf("unexpected Content-Type header: %v", got)
	}

	encodedCredentials := url.QueryEscape("client id") + ":" + url.QueryEscape("s:e/c ret")
	wantAuthorization := "Basic " + base64.StdEncoding.EncodeToString([]byte(encodedCredentials))
	if got := captured.Headers["Authorization"]; !reflect.DeepEqual(
		got,
		[]string{wantAuthorization},
	) {
		t.Fatalf("unexpected Authorization header: %v", got)
	}

	form, err := url.ParseQuery(string(captured.Body))
	if err != nil {
		t.Fatalf("parse form: %v", err)
	}
	if got := form.Get("grant_type"); got != "client_credentials" {
		t.Fatalf("unexpected grant_type: %q", got)
	}
	if got := form.Get("scope"); got != "users:read projects:read" {
		t.Fatalf("unexpected scope: %q", got)
	}
	if got := form["resource"]; !reflect.DeepEqual(got, []string{"one", "two"}) {
		t.Fatalf("unexpected repeated parameter: %v", got)
	}
	if form.Has("client_id") || form.Has("client_secret") {
		t.Fatalf("Basic credentials leaked into form: %v", form)
	}

	if token.ExpiresIn != time.Minute || !token.ExpiresAt.Equal(now.Add(time.Minute)) {
		t.Fatalf("unexpected expiry: %#v", token)
	}
	if token.Scope != "users:read projects:read" {
		t.Fatalf("expected effective request scope fallback, got %q", token.Scope)
	}
	number, ok := token.Extra["provider_number"].(interface{ String() string })
	if !ok || number.String() != "1.25" {
		t.Fatalf("expected json.Number fidelity, got %#v", token.Extra["provider_number"])
	}
}

func TestExecutorClientCredentialsPostSuppressesAmbientAuthorization(t *testing.T) {
	var captured *ferrethttp.Request
	httpClient := newRecordingTokenHTTPClient(
		func(_ context.Context, request *ferrethttp.Request) (*ferrethttp.Response, error) {
			captured = request

			return successfulTokenResponse(`{"access_token":"access","token_type":"Bearer"}`), nil
		},
	)
	executor, err := NewExecutor(httpClient)
	if err != nil {
		t.Fatalf("construct executor: %v", err)
	}

	_, err = executor.ClientCredentials(
		context.Background(),
		testOAuthClient(ClientAuthMethodPost, "client", "secret"),
		ClientCredentialsOptions{},
	)
	if err != nil {
		t.Fatalf("client credentials: %v", err)
	}

	values, exists := captured.Headers["Authorization"]
	if !exists || values != nil {
		t.Fatalf("expected explicit nil Authorization suppression marker, got %#v", captured.Headers)
	}

	form, err := url.ParseQuery(string(captured.Body))
	if err != nil {
		t.Fatalf("parse form: %v", err)
	}
	if form.Get("client_id") != "client" || form.Get("client_secret") != "secret" {
		t.Fatalf("POST credentials missing from form: %v", form)
	}
}

func TestExecutorClientCredentialsRejectsPublicClient(t *testing.T) {
	called := false
	httpClient := newRecordingTokenHTTPClient(
		func(context.Context, *ferrethttp.Request) (*ferrethttp.Response, error) {
			called = true

			return nil, nil
		},
	)
	executor, err := NewExecutor(httpClient)
	if err != nil {
		t.Fatalf("construct executor: %v", err)
	}

	_, err = executor.ClientCredentials(
		context.Background(),
		testOAuthClient(ClientAuthMethodNone, "public", ""),
		ClientCredentialsOptions{},
	)
	if !errors.Is(err, ErrInvalidGrant) ||
		!strings.Contains(err.Error(), "authenticated client") {
		t.Fatalf("expected authenticated-client error, got %v", err)
	}
	if called {
		t.Fatal("invalid public client reached HTTP transport")
	}
}

func TestExecutorTokenExtensionGrant(t *testing.T) {
	var captured *ferrethttp.Request
	httpClient := newRecordingTokenHTTPClient(
		func(_ context.Context, request *ferrethttp.Request) (*ferrethttp.Response, error) {
			captured = request

			return successfulTokenResponse(`{"access_token":"access","token_type":"mac"}`), nil
		},
	)
	executor, err := NewExecutor(httpClient)
	if err != nil {
		t.Fatalf("construct executor: %v", err)
	}

	token, err := executor.Token(
		context.Background(),
		testOAuthClient(ClientAuthMethodNone, "public", ""),
		TokenOptions{Parameters: Parameters{
			"grant_type": {
				"urn:ietf:params:oauth:grant-type:jwt-bearer",
			},
			"assertion": {"signed-assertion"},
			"scope":     {"users:read"},
		}},
	)
	if err != nil {
		t.Fatalf("extension grant: %v", err)
	}
	if token.TokenType != "mac" || token.Scope != "users:read" {
		t.Fatalf("unexpected token: %#v", token)
	}

	values, exists := captured.Headers["Authorization"]
	if !exists || values != nil {
		t.Fatalf("expected Authorization suppression marker, got %#v", captured.Headers)
	}
	form, err := url.ParseQuery(string(captured.Body))
	if err != nil {
		t.Fatalf("parse form: %v", err)
	}
	if form.Get("client_id") != "public" ||
		form.Get("assertion") != "signed-assertion" {
		t.Fatalf("unexpected extension form: %v", form)
	}
}

func TestExecutorTokenRequiresOneAbsoluteURIGrantType(t *testing.T) {
	tests := []struct {
		name       string
		grantTypes []string
	}{
		{name: "missing"},
		{name: "empty", grantTypes: []string{""}},
		{name: "registered short name", grantTypes: []string{"password"}},
		{name: "fragment", grantTypes: []string{"urn:example:grant#fragment"}},
		{name: "repeated", grantTypes: []string{"urn:one", "urn:two"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			httpClient := newRecordingTokenHTTPClient(
				func(context.Context, *ferrethttp.Request) (*ferrethttp.Response, error) {
					t.Fatal("invalid grant type reached HTTP transport")

					return nil, nil
				},
			)
			executor, err := NewExecutor(httpClient)
			if err != nil {
				t.Fatalf("construct executor: %v", err)
			}

			parameters := make(Parameters)
			if test.grantTypes != nil {
				parameters["grant_type"] = test.grantTypes
			}

			_, err = executor.Token(
				context.Background(),
				testOAuthClient(ClientAuthMethodNone, "public", ""),
				TokenOptions{Parameters: parameters},
			)
			if !errors.Is(err, ErrInvalidGrant) {
				t.Fatalf("expected ErrInvalidGrant, got %v", err)
			}
		})
	}
}

func TestExecutorRejectsReservedParameterCollisions(t *testing.T) {
	tests := []struct {
		run  func(*Executor) error
		name string
	}{
		{
			name: "client credentials scope",
			run: func(executor *Executor) error {
				_, err := executor.ClientCredentials(
					context.Background(),
					testOAuthClient(ClientAuthMethodBasic, "client", "secret"),
					ClientCredentialsOptions{Parameters: Parameters{"scope": {"override"}}},
				)

				return err
			},
		},
		{
			name: "refresh token",
			run: func(executor *Executor) error {
				_, err := executor.Refresh(
					context.Background(),
					testOAuthClient(ClientAuthMethodBasic, "client", "secret"),
					nil,
					RefreshOptions{
						RefreshToken: "refresh",
						Parameters:   Parameters{"refresh_token": {"override"}},
					},
				)

				return err
			},
		},
		{
			name: "extension client assertion wildcard",
			run: func(executor *Executor) error {
				_, err := executor.Token(
					context.Background(),
					testOAuthClient(ClientAuthMethodBasic, "client", "secret"),
					TokenOptions{Parameters: Parameters{
						"grant_type":              {"urn:example:grant"},
						"client_assertion_custom": {"override"},
					}},
				)

				return err
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			httpClient := newRecordingTokenHTTPClient(
				func(context.Context, *ferrethttp.Request) (*ferrethttp.Response, error) {
					t.Fatal("reserved collision reached HTTP transport")

					return nil, nil
				},
			)
			executor, err := NewExecutor(httpClient)
			if err != nil {
				t.Fatalf("construct executor: %v", err)
			}

			err = test.run(executor)
			if !errors.Is(err, ErrInvalidGrant) ||
				!strings.Contains(err.Error(), "reserved") {
				t.Fatalf("expected reserved-parameter error, got %v", err)
			}
		})
	}
}

func TestExecutorRefreshRetainsAndRotatesRefreshToken(t *testing.T) {
	responses := []string{
		`{"access_token":"access-1","token_type":"Bearer"}`,
		`{"access_token":"access-2","token_type":"Bearer","refresh_token":"rotated"}`,
	}
	call := 0
	httpClient := newRecordingTokenHTTPClient(
		func(_ context.Context, request *ferrethttp.Request) (*ferrethttp.Response, error) {
			form, err := url.ParseQuery(string(request.Body))
			if err != nil {
				t.Fatalf("parse form: %v", err)
			}
			if form.Get("refresh_token") != "original" {
				t.Fatalf("unexpected submitted refresh token: %v", form)
			}

			response := successfulTokenResponse(responses[call])
			call++

			return response, nil
		},
	)
	executor, err := NewExecutor(httpClient)
	if err != nil {
		t.Fatalf("construct executor: %v", err)
	}
	previous := &TokenSet{
		AccessToken:  "old",
		TokenType:    "Bearer",
		RefreshToken: "original",
		Scope:        "read write",
	}

	retained, err := executor.Refresh(
		context.Background(),
		testOAuthClient(ClientAuthMethodBasic, "client", "secret"),
		previous,
		RefreshOptions{Scope: []string{"read"}},
	)
	if err != nil {
		t.Fatalf("refresh without rotation: %v", err)
	}
	if retained.RefreshToken != "original" || retained.Scope != "read" {
		t.Fatalf("refresh fallback failed: %#v", retained)
	}

	rotated, err := executor.Refresh(
		context.Background(),
		testOAuthClient(ClientAuthMethodBasic, "client", "secret"),
		previous,
		RefreshOptions{},
	)
	if err != nil {
		t.Fatalf("refresh with rotation: %v", err)
	}
	if rotated.RefreshToken != "rotated" || rotated.Scope != "read write" {
		t.Fatalf("refresh rotation/fallback failed: %#v", rotated)
	}
}

func TestExecutorRefreshRejectsScopeEscalation(t *testing.T) {
	httpClient := newRecordingTokenHTTPClient(
		func(context.Context, *ferrethttp.Request) (*ferrethttp.Response, error) {
			t.Fatal("scope escalation reached HTTP transport")

			return nil, nil
		},
	)
	executor, err := NewExecutor(httpClient)
	if err != nil {
		t.Fatalf("construct executor: %v", err)
	}

	_, err = executor.Refresh(
		context.Background(),
		testOAuthClient(ClientAuthMethodBasic, "client", "secret"),
		&TokenSet{RefreshToken: "refresh", Scope: "read"},
		RefreshOptions{Scope: []string{"read", "write"}},
	)
	if !errors.Is(err, ErrInvalidGrant) || !strings.Contains(err.Error(), "previous") {
		t.Fatalf("expected scope escalation error, got %v", err)
	}
}

func TestExecutorSanitizesOAuthErrorSecrets(t *testing.T) {
	const (
		clientSecret = "client secret"
		assertion    = "signed.assertion"
	)
	httpClient := newRecordingTokenHTTPClient(
		func(_ context.Context, request *ferrethttp.Request) (*ferrethttp.Response, error) {
			authorization := request.Headers["Authorization"][0]

			return &ferrethttp.Response{
				StatusCode: 400,
				Headers: ferrethttp.Headers{
					"Content-Type": {"application/json"},
				},
				Body: []byte(`{
						"error":"invalid_grant",
						"error_description":"bad ` + clientSecret + ` ` + assertion + ` ` + authorization + `",
						"error_uri":"https://example.com/error?assertion=` + url.QueryEscape(assertion) + `"
					}`),
			}, nil
		},
	)
	executor, err := NewExecutor(httpClient)
	if err != nil {
		t.Fatalf("construct executor: %v", err)
	}

	_, err = executor.Token(
		context.Background(),
		testOAuthClient(ClientAuthMethodBasic, "client", clientSecret),
		TokenOptions{Parameters: Parameters{
			"grant_type": {"urn:example:grant"},
			"assertion":  {assertion},
		}},
	)

	var oauthError *Error
	if !errors.As(err, &oauthError) {
		t.Fatalf("expected typed OAuth error, got %T: %v", err, err)
	}
	if oauthError.Code != "invalid_grant" || oauthError.StatusCode != 400 {
		t.Fatalf("unexpected typed OAuth error: %#v", oauthError)
	}

	formatted := oauthError.Error() + oauthError.URI
	for _, secret := range []string{clientSecret, assertion, url.QueryEscape(assertion)} {
		if strings.Contains(formatted, secret) {
			t.Fatalf("error leaked secret %q: %s", secret, formatted)
		}
	}
}

func TestExecutorPreservesTransportErrorChainAndSanitizesMessage(t *testing.T) {
	sentinel := errors.New("transport saw signed-assertion")
	httpClient := newRecordingTokenHTTPClient(
		func(context.Context, *ferrethttp.Request) (*ferrethttp.Response, error) {
			return nil, sentinel
		},
	)
	executor, err := NewExecutor(httpClient)
	if err != nil {
		t.Fatalf("construct executor: %v", err)
	}

	_, err = executor.Token(
		context.Background(),
		testOAuthClient(ClientAuthMethodNone, "client", ""),
		TokenOptions{Parameters: Parameters{
			"grant_type": {"urn:example:grant"},
			"assertion":  {"signed-assertion"},
		}},
	)
	if !errors.Is(err, sentinel) {
		t.Fatalf("transport error chain was lost: %v", err)
	}
	if strings.Contains(err.Error(), "signed-assertion") {
		t.Fatalf("transport error leaked assertion: %v", err)
	}
}

func TestExecutorAppliesShorterContextDeadline(t *testing.T) {
	httpClient := newRecordingTokenHTTPClient(
		func(ctx context.Context, _ *ferrethttp.Request) (*ferrethttp.Response, error) {
			<-ctx.Done()

			return nil, ctx.Err()
		},
	)
	executor, err := NewExecutor(httpClient)
	if err != nil {
		t.Fatalf("construct executor: %v", err)
	}

	_, err = executor.Token(
		context.Background(),
		testOAuthClient(ClientAuthMethodNone, "client", ""),
		TokenOptions{
			Parameters: Parameters{"grant_type": {"urn:example:grant"}},
			Timeout:    time.Millisecond,
		},
	)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline error chain, got %v", err)
	}
}

func successfulTokenResponse(body string) *ferrethttp.Response {
	return &ferrethttp.Response{
		StatusCode: 200,
		Headers: ferrethttp.Headers{
			"Content-Type": {"application/json; charset=utf-8"},
		},
		Body: []byte(body),
	}
}

func testOAuthClient(
	authMethod ClientAuthMethod,
	clientID string,
	clientSecret string,
) *Client {
	return &Client{
		Provider: &Provider{
			TokenEndpoint: "https://auth.example.com/token",
		},
		ClientID:     clientID,
		ClientSecret: clientSecret,
		AuthMethod:   authMethod,
	}
}
