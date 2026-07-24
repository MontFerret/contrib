package core

import (
	"context"
	"errors"
	"io"
	stdhttp "net/http"
	"strings"
	"testing"

	ferrethttp "github.com/MontFerret/ferret/v2/pkg/net/http"
)

func TestExecutorSuppressionMarkerPreventsAmbientAuthorization(t *testing.T) {
	var authorization []string
	httpClient, err := ferrethttp.NewWithTransport(
		tokenRoundTripFunc(func(request *stdhttp.Request) (*stdhttp.Response, error) {
			authorization = append([]string(nil), request.Header["Authorization"]...)

			return &stdhttp.Response{
				StatusCode: 200,
				Status:     "200 OK",
				Header:     stdhttp.Header{"Content-Type": {"application/json"}},
				Body: io.NopCloser(strings.NewReader(
					`{"access_token":"access","token_type":"Bearer"}`,
				)),
				Request: request,
			}, nil
		}),
		ferrethttp.WithDefaultHeader("Authorization", "Bearer ambient-secret"),
	)
	if err != nil {
		t.Fatalf("construct policy client: %v", err)
	}
	executor, err := NewExecutor(httpClient)
	if err != nil {
		t.Fatalf("construct executor: %v", err)
	}

	_, err = executor.Token(
		context.Background(),
		testOAuthClient(ClientAuthMethodNone, "public", ""),
		TokenOptions{Parameters: Parameters{"grant_type": {"urn:example:grant"}}},
	)
	if err != nil {
		t.Fatalf("extension grant: %v", err)
	}
	if authorization != nil {
		t.Fatalf("ambient Authorization reached transport: %v", authorization)
	}
}

func TestExecutorSuppressionMarkerPreservesBlockedHeaderPolicyError(t *testing.T) {
	called := false
	httpClient, err := ferrethttp.NewWithTransport(
		tokenRoundTripFunc(func(request *stdhttp.Request) (*stdhttp.Response, error) {
			called = true

			return nil, errors.New("unexpected transport call")
		}),
		ferrethttp.WithBlockedRequestHeaders("Authorization"),
	)
	if err != nil {
		t.Fatalf("construct policy client: %v", err)
	}
	executor, err := NewExecutor(httpClient)
	if err != nil {
		t.Fatalf("construct executor: %v", err)
	}

	_, err = executor.Token(
		context.Background(),
		testOAuthClient(ClientAuthMethodNone, "public", ""),
		TokenOptions{Parameters: Parameters{"grant_type": {"urn:example:grant"}}},
	)

	var policyError *ferrethttp.PolicyError
	if !errors.As(err, &policyError) {
		t.Fatalf("expected preserved PolicyError, got %T: %v", err, err)
	}
	if called {
		t.Fatal("blocked Authorization marker reached transport")
	}
}

func TestExecutorPreservesResponseSizePolicyError(t *testing.T) {
	t.Parallel()

	httpClient, err := ferrethttp.NewWithTransport(
		tokenRoundTripFunc(func(request *stdhttp.Request) (*stdhttp.Response, error) {
			return &stdhttp.Response{
				StatusCode: 200,
				Status:     "200 OK",
				Header:     stdhttp.Header{"Content-Type": {"application/json"}},
				Body: io.NopCloser(strings.NewReader(
					`{"access_token":"a-token-that-exceeds-the-limit","token_type":"Bearer"}`,
				)),
				Request: request,
			}, nil
		}),
		ferrethttp.WithMaxResponseSize(32),
	)
	if err != nil {
		t.Fatalf("construct policy client: %v", err)
	}

	executor, err := NewExecutor(httpClient)
	if err != nil {
		t.Fatalf("construct executor: %v", err)
	}

	_, err = executor.ClientCredentials(
		context.Background(),
		testOAuthClient(ClientAuthMethodBasic, "client", "secret"),
		ClientCredentialsOptions{},
	)

	var limitError *ferrethttp.ResponseBodyLimitError
	if !errors.As(err, &limitError) {
		t.Fatalf("expected preserved ResponseBodyLimitError, got %T: %v", err, err)
	}
	if limitError.Limit != 32 {
		t.Fatalf("response limit = %d, want 32", limitError.Limit)
	}
}

func TestExecutorUsesHostRedirectPolicy(t *testing.T) {
	t.Parallel()

	newTransport := func(calls *[]string) tokenRoundTripFunc {
		return func(request *stdhttp.Request) (*stdhttp.Response, error) {
			*calls = append(*calls, request.URL.Path)

			if request.URL.Path == "/token" {
				return &stdhttp.Response{
					StatusCode: stdhttp.StatusFound,
					Status:     "302 Found",
					Header: stdhttp.Header{
						"Location": {"https://auth.example.com/final"},
					},
					Body:    io.NopCloser(strings.NewReader("")),
					Request: request,
				}, nil
			}

			return &stdhttp.Response{
				StatusCode: 200,
				Status:     "200 OK",
				Header:     stdhttp.Header{"Content-Type": {"application/json"}},
				Body: io.NopCloser(strings.NewReader(
					`{"access_token":"access","token_type":"Bearer"}`,
				)),
				Request: request,
			}, nil
		}
	}

	var blockedCalls []string
	blockedClient, err := ferrethttp.NewWithTransport(
		newTransport(&blockedCalls),
		ferrethttp.WithFollowRedirects(false),
	)
	if err != nil {
		t.Fatalf("construct no-redirect client: %v", err)
	}
	blockedExecutor, err := NewExecutor(blockedClient)
	if err != nil {
		t.Fatalf("construct no-redirect executor: %v", err)
	}

	_, err = blockedExecutor.ClientCredentials(
		context.Background(),
		testOAuthClient(ClientAuthMethodBasic, "client", "secret"),
		ClientCredentialsOptions{},
	)
	if !errors.Is(err, ErrInvalidTokenResponse) {
		t.Fatalf("expected redirect response rejection, got %v", err)
	}
	if len(blockedCalls) != 1 || blockedCalls[0] != "/token" {
		t.Fatalf("disabled redirect calls = %v", blockedCalls)
	}

	var allowedCalls []string
	allowedClient, err := ferrethttp.NewWithTransport(newTransport(&allowedCalls))
	if err != nil {
		t.Fatalf("construct redirecting client: %v", err)
	}
	allowedExecutor, err := NewExecutor(allowedClient)
	if err != nil {
		t.Fatalf("construct redirecting executor: %v", err)
	}

	token, err := allowedExecutor.ClientCredentials(
		context.Background(),
		testOAuthClient(ClientAuthMethodBasic, "client", "secret"),
		ClientCredentialsOptions{},
	)
	if err != nil {
		t.Fatalf("follow redirect token request: %v", err)
	}
	if token.AccessToken != "access" {
		t.Fatalf("redirected access token = %q", token.AccessToken)
	}
	if len(allowedCalls) != 2 || allowedCalls[0] != "/token" || allowedCalls[1] != "/final" {
		t.Fatalf("enabled redirect calls = %v", allowedCalls)
	}
}
