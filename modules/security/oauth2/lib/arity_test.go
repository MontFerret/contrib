package lib

import (
	"context"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestFunctionsValidateArity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		function runtime.Function
		name     string
		min      int
		max      int
	}{
		{name: "DISCOVER", function: Discover, min: 1, max: 2},
		{name: "PROVIDER", function: Provider, min: 1, max: 1},
		{name: "CLIENT", function: Client, min: 2, max: 2},
		{name: "CLIENT_CREDENTIALS", function: ClientCredentials, min: 1, max: 2},
		{name: "REFRESH", function: Refresh, min: 2, max: 2},
		{name: "TOKEN", function: Token, min: 2, max: 2},
		{name: "ACCESS_TOKEN", function: AccessToken, min: 1, max: 1},
		{name: "REFRESH_TOKEN", function: RefreshToken, min: 1, max: 1},
		{name: "ID_TOKEN", function: IDToken, min: 1, max: 1},
		{name: "TOKEN_TYPE", function: TokenType, min: 1, max: 1},
		{name: "SCOPES", function: Scopes, min: 1, max: 1},
		{name: "EXPIRES_AT", function: ExpiresAt, min: 1, max: 1},
		{name: "EXPIRED", function: Expired, min: 1, max: 2},
		{name: "VALID_FOR", function: ValidFor, min: 1, max: 1},
		{name: "AUTH_HEADER", function: AuthHeader, min: 1, max: 1},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			tooFew := make([]runtime.Value, test.min-1)
			if _, err := test.function(context.Background(), tooFew...); err == nil {
				t.Fatal("expected too-few-arguments error")
			}

			tooMany := make([]runtime.Value, test.max+1)
			if _, err := test.function(context.Background(), tooMany...); err == nil {
				t.Fatal("expected too-many-arguments error")
			}
		})
	}
}

func TestFunctionsRejectWrongHostAndConfigTypes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	stringValue := runtime.NewString("wrong")

	if _, err := Provider(ctx, stringValue); err == nil {
		t.Fatal("PROVIDER accepted a string config")
	}
	if _, err := Client(ctx, stringValue, runtime.NewObject()); err == nil {
		t.Fatal("CLIENT accepted a non-provider host value")
	}
	if _, err := AccessToken(ctx, stringValue); err == nil {
		t.Fatal("ACCESS_TOKEN accepted a non-token host value")
	}
	if _, err := Discover(ctx, runtime.NewInt(1)); err == nil {
		t.Fatal("DISCOVER accepted a non-string issuer")
	}
}
