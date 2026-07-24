package lib

import (
	"fmt"

	"github.com/MontFerret/contrib/modules/security/oauth2/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func providerArgument(value runtime.Value) (*core.Provider, error) {
	provider, ok := value.(*providerValue)
	if !ok || provider == nil || provider.target == nil {
		return nil, fmt.Errorf("expected an OAuth2 provider")
	}

	return provider.provider(), nil
}

func clientArgument(value runtime.Value) (*core.Client, error) {
	client, ok := value.(*clientValue)
	if !ok || client == nil || client.target == nil {
		return nil, fmt.Errorf("expected an OAuth2 client")
	}

	return client.client(), nil
}

func tokenArgument(value runtime.Value) (*core.TokenSet, *tokenValue, error) {
	token, ok := value.(*tokenValue)
	if !ok || token == nil || token.target == nil {
		return nil, nil, fmt.Errorf("expected an OAuth2 token set")
	}

	return token.token(), token, nil
}
