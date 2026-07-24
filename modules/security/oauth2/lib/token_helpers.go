package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/security/oauth2/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// AccessToken intentionally returns the raw access token.
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

// RefreshToken intentionally returns the raw refresh token, or NONE.
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

// IDToken intentionally returns the raw ID token, or NONE.
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

// ExpiresAt returns the token expiration timestamp, or NONE.
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
