package oauth2

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/MontFerret/ferret/v2"
	ferretnet "github.com/MontFerret/ferret/v2/pkg/net"
	ferrethttp "github.com/MontFerret/ferret/v2/pkg/net/http"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk/sdktest"
)

func TestNewSmoke(t *testing.T) {
	t.Parallel()

	module := New()
	if module == nil {
		t.Fatal("expected module")
	}
	if module.Name() != "security/oauth2" {
		t.Fatalf("module name = %q", module.Name())
	}
}

func TestModuleRunsManualClientCredentialsFlow(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		if request.Method != http.MethodPost || request.URL.Path != "/token" {
			http.Error(writer, "unexpected request", http.StatusNotFound)

			return
		}

		clientID, clientSecret, ok := request.BasicAuth()
		if !ok || clientID != "test-client" || clientSecret != "test-secret" {
			http.Error(writer, "invalid client authentication", http.StatusUnauthorized)

			return
		}

		rawBody, err := io.ReadAll(request.Body)
		if err != nil {
			http.Error(writer, "read form", http.StatusBadRequest)

			return
		}

		body, err := url.ParseQuery(string(rawBody))
		if err != nil {
			http.Error(writer, "invalid form", http.StatusBadRequest)

			return
		}
		if body.Get("grant_type") != "client_credentials" ||
			body.Get("scope") != "users:read" {
			http.Error(writer, "invalid grant", http.StatusBadRequest)

			return
		}

		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(map[string]any{
			"access_token":  "module-access-token",
			"token_type":    "Bearer",
			"refresh_token": "module-refresh-token",
			"scope":         "users:read",
		})
	}))
	defer server.Close()

	network, err := ferretnet.New(
		ferretnet.WithHTTPPolicies(ferrethttp.WithAllowLocalhost(true)),
	)
	if err != nil {
		t.Fatalf("construct network: %v", err)
	}

	harness := sdktest.New(
		t,
		ferret.WithModules(New()),
		ferret.WithNetwork(network),
		ferret.WithRuntimeParam("tokenEndpoint", runtime.NewString(server.URL+"/token")),
	)

	output, err := harness.Run(context.Background(), `
		LET provider = SECURITY::OAUTH2::PROVIDER({
			tokenEndpoint: @tokenEndpoint,
			insecureAllowHTTP: true
		})
		LET client = SECURITY::OAUTH2::CLIENT(provider, {
			clientID: "test-client",
			clientSecret: "test-secret"
		})
		LET token = SECURITY::OAUTH2::CLIENT_CREDENTIALS(client, {
			scope: ["users:read"]
		})
		RETURN {
			tokenType: token.tokenType,
			scopes: token.scopes,
			hasRefreshToken: token.hasRefreshToken,
			header: SECURITY::OAUTH2::AUTH_HEADER(token)
		}
	`)
	if err != nil {
		t.Fatalf("run OAuth2 flow: %v", err)
	}

	var result struct {
		Header          map[string]string `json:"header"`
		TokenType       string            `json:"tokenType"`
		Scopes          []string          `json:"scopes"`
		HasRefreshToken bool              `json:"hasRefreshToken"`
	}
	if err := json.Unmarshal(output.Content, &result); err != nil {
		t.Fatalf("decode result: %v", err)
	}
	if result.TokenType != "Bearer" {
		t.Fatalf("tokenType = %q", result.TokenType)
	}
	if len(result.Scopes) != 1 || result.Scopes[0] != "users:read" {
		t.Fatalf("scopes = %v", result.Scopes)
	}
	if !result.HasRefreshToken {
		t.Fatal("expected refresh-token presence")
	}
	if result.Header["Authorization"] != "Bearer module-access-token" {
		t.Fatalf("Authorization = %q", result.Header["Authorization"])
	}
}
