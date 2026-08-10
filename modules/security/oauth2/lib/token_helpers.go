package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/security/oauth2/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// AccessToken intentionally returns the raw access token.
//
// The returned secret is no longer protected by the token handle's redacted
// formatting and serialization.
//
// @param token {OAuthToken} Token set.
// @return {String} Raw access token.
func AccessToken(_ context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 1); err != nil {
		return runtime.None, err
	}

	token, _, err := tokenArgument(args[0])
	if err != nil {
		return runtime.None, err
	}

	return runtime.NewString(token.AccessToken), nil
}

// RefreshToken intentionally returns the raw refresh token when present.
//
// The returned secret is no longer protected by the token handle's redacted
// formatting and serialization.
//
// @param token {OAuthToken} Token set.
// @return {String|None} Raw refresh token or None.
func RefreshToken(_ context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 1); err != nil {
		return runtime.None, err
	}

	token, _, err := tokenArgument(args[0])
	if err != nil {
		return runtime.None, err
	}

	if token.RefreshToken == "" {
		return runtime.None, nil
	}

	return runtime.NewString(token.RefreshToken), nil
}

// IDToken intentionally returns the raw ID token when present.
//
// The token is returned without parsing or validation and is no longer
// protected by the token handle's redacted formatting and serialization.
//
// @param token {OAuthToken} Token set.
// @return {String|None} Raw ID token or None.
func IDToken(_ context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 1); err != nil {
		return runtime.None, err
	}

	token, _, err := tokenArgument(args[0])
	if err != nil {
		return runtime.None, err
	}

	if token.IDToken == "" {
		return runtime.None, nil
	}

	return runtime.NewString(token.IDToken), nil
}

// TokenType returns the provider-supplied token type.
//
// @param token {OAuthToken} Token set.
// @return {String} Provider-supplied token type.
func TokenType(_ context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 1); err != nil {
		return runtime.None, err
	}

	token, _, err := tokenArgument(args[0])
	if err != nil {
		return runtime.None, err
	}

	return runtime.NewString(token.TokenType), nil
}

// Scopes returns token scopes split on ASCII spaces.
//
// @param token {OAuthToken} Token set.
// @return {Array<String>} Token scopes.
func Scopes(_ context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 1); err != nil {
		return runtime.None, err
	}

	token, _, err := tokenArgument(args[0])
	if err != nil {
		return runtime.None, err
	}

	return stringArray(token.Scopes()), nil
}

// ExpiresAt returns the token expiration timestamp when known.
//
// @param token {OAuthToken} Token set.
// @return {DateTime|None} Expiration timestamp or None.
func ExpiresAt(_ context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 1); err != nil {
		return runtime.None, err
	}

	token, _, err := tokenArgument(args[0])
	if err != nil {
		return runtime.None, err
	}

	if token.ExpiresAt.IsZero() {
		return runtime.None, nil
	}

	return runtime.NewDateTime(token.ExpiresAt), nil
}

// Expired reports whether a token is expired with an optional safety skew.
//
// Tokens with unknown expiration are not considered expired.
//
// @param token {OAuthToken} Token set.
// @param options {Object?} Expiration safety-skew options.
// @return {Boolean} Whether the token is expired.
func Expired(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 2); err != nil {
		return runtime.None, err
	}

	token, wrapped, err := tokenArgument(args[0])
	if err != nil {
		return runtime.None, err
	}

	var optionsValue runtime.Value

	if len(args) == 2 {
		optionsValue = args[1]
	}

	skew, err := decodeExpiredOptions(ctx, optionsValue)
	if err != nil {
		return runtime.None, err
	}

	return runtime.NewBoolean(token.Expired(wrapped.now(), skew)), nil
}

// ValidFor returns remaining token lifetime in rounded-up milliseconds.
//
// @param token {OAuthToken} Token set.
// @return {Int|None} Remaining milliseconds or None when expiration is unknown.
func ValidFor(_ context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 1); err != nil {
		return runtime.None, err
	}

	token, wrapped, err := tokenArgument(args[0])
	if err != nil {
		return runtime.None, err
	}

	remaining, known := token.ValidFor(wrapped.now())
	if !known {
		return runtime.None, nil
	}

	return runtime.NewInt64(ceilMilliseconds(remaining)), nil
}

// AuthHeader intentionally materializes a Bearer Authorization header.
//
// The returned object contains the raw access token and should be passed
// directly to an HTTP client rather than logged or serialized.
//
// @param token {OAuthToken} Bearer token set.
// @return {Object} Authorization header object.
func AuthHeader(_ context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 1); err != nil {
		return runtime.None, err
	}

	token, _, err := tokenArgument(args[0])
	if err != nil {
		return runtime.None, err
	}

	header, err := core.AuthorizationHeader(token)
	if err != nil {
		return runtime.None, err
	}

	return runtime.NewObjectWith(map[string]runtime.Value{
		"Authorization": runtime.NewString(header["Authorization"]),
	}), nil
}
