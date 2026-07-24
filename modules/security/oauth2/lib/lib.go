package lib

import (
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// RegisterLib atomically registers the SECURITY::OAUTH2 namespace functions.
func RegisterLib(ns runtime.Namespace) error {
	return sdk.RegisterFunctions(
		ns,
		sdk.Func("DISCOVER", Discover),
		sdk.Func("PROVIDER", Provider),
		sdk.Func("CLIENT", Client),
		sdk.Func("CLIENT_CREDENTIALS", ClientCredentials),
		sdk.Func("REFRESH", Refresh),
		sdk.Func("TOKEN", Token),
		sdk.Func("ACCESS_TOKEN", AccessToken),
		sdk.Func("REFRESH_TOKEN", RefreshToken),
		sdk.Func("ID_TOKEN", IDToken),
		sdk.Func("TOKEN_TYPE", TokenType),
		sdk.Func("SCOPES", Scopes),
		sdk.Func("EXPIRES_AT", ExpiresAt),
		sdk.Func("EXPIRED", Expired),
		sdk.Func("VALID_FOR", ValidFor),
		sdk.Func("AUTH_HEADER", AuthHeader),
	)
}
