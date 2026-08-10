package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/security/oauth2/core"
	ferretnet "github.com/MontFerret/ferret/v2/pkg/net"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// ClientCredentials performs an OAuth client_credentials grant.
//
// @param client {OAuthClient} OAuth client handle.
// @param options {Object?} Scope, audience, parameters, and timeout.
// @return {OAuthToken} Acquired token set.
func ClientCredentials(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 2); err != nil {
		return runtime.None, err
	}

	client, err := clientArgument(args[0])
	if err != nil {
		return runtime.None, err
	}

	var optionsValue runtime.Value
	if len(args) == 2 {
		optionsValue = args[1]
	}

	options, err := decodeClientCredentialsOptions(ctx, optionsValue)
	if err != nil {
		return runtime.None, err
	}

	executor, err := executorFrom(ctx)
	if err != nil {
		return runtime.None, err
	}

	token, err := executor.ClientCredentials(ctx, client, options)
	if err != nil {
		return runtime.None, err
	}

	return newTokenValue(token), nil
}

// Refresh performs an OAuth refresh_token grant.
//
// @param client {OAuthClient} OAuth client handle.
// @param tokenOrOptions {OAuthToken|Object} Existing token set or refresh-token options.
// @return {OAuthToken} Refreshed token set.
func Refresh(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 2, 2); err != nil {
		return runtime.None, err
	}

	client, err := clientArgument(args[0])
	if err != nil {
		return runtime.None, err
	}

	var previous *core.TokenSet
	var options core.RefreshOptions

	if token, _, tokenErr := tokenArgument(args[1]); tokenErr == nil {
		previous = token
	} else {
		options, err = decodeRefreshOptions(ctx, args[1])
		if err != nil {
			return runtime.None, err
		}
	}

	executor, err := executorFrom(ctx)
	if err != nil {
		return runtime.None, err
	}

	token, err := executor.Refresh(ctx, client, previous, options)
	if err != nil {
		return runtime.None, err
	}

	return newTokenValue(token), nil
}

// Token performs an OAuth extension grant.
//
// @param client {OAuthClient} OAuth client handle.
// @param parameters {Object} Extension-grant parameters.
// @return {OAuthToken} Acquired token set.
func Token(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 2, 2); err != nil {
		return runtime.None, err
	}

	client, err := clientArgument(args[0])
	if err != nil {
		return runtime.None, err
	}

	parameters, err := decodeParameters(ctx, args[1], "parameters")
	if err != nil {
		return runtime.None, err
	}

	executor, err := executorFrom(ctx)
	if err != nil {
		return runtime.None, err
	}

	token, err := executor.Token(ctx, client, core.TokenOptions{Parameters: parameters})
	if err != nil {
		return runtime.None, err
	}

	return newTokenValue(token), nil
}

func executorFrom(ctx context.Context) (*core.Executor, error) {
	httpClient, err := ferretnet.HTTPClientFrom(ctx)
	if err != nil {
		return nil, err
	}

	return core.NewExecutor(httpClient)
}
