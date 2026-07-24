package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/security/oauth2/core"
	ferretnet "github.com/MontFerret/ferret/v2/pkg/net"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// Discover loads OAuth authorization-server metadata for an issuer.
func Discover(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 2); err != nil {
		return runtime.None, err
	}

	issuer, err := sdk.DecodeArg[string](ctx, args, 0, sdk.RequireType(runtime.TypeString))
	if err != nil {
		return runtime.None, err
	}

	var optionsValue runtime.Value
	if len(args) == 2 {
		optionsValue = args[1]
	}

	options, err := decodeDiscoveryOptions(ctx, optionsValue)
	if err != nil {
		return runtime.None, err
	}

	httpClient, err := ferretnet.HTTPClientFrom(ctx)
	if err != nil {
		return runtime.None, err
	}

	provider, err := core.Discover(ctx, httpClient, issuer, options)
	if err != nil {
		return runtime.None, err
	}

	return newProviderValue(provider), nil
}

// Provider constructs a manually configured OAuth provider.
func Provider(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 1); err != nil {
		return runtime.None, err
	}

	config, err := decodeProviderConfig(ctx, args[0])
	if err != nil {
		return runtime.None, err
	}

	provider, err := core.NewProvider(config)
	if err != nil {
		return runtime.None, err
	}

	return newProviderValue(provider), nil
}

// Client constructs an OAuth client for a provider.
func Client(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 2, 2); err != nil {
		return runtime.None, err
	}

	provider, err := providerArgument(args[0])
	if err != nil {
		return runtime.None, err
	}

	config, err := decodeClientConfig(ctx, args[1])
	if err != nil {
		return runtime.None, err
	}

	client, err := core.NewClient(provider, config)
	if err != nil {
		return runtime.None, err
	}

	return newClientValue(client), nil
}
