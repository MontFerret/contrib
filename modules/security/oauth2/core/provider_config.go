package core

import "fmt"

// ProviderConfig configures an OAuth authorization server.
type ProviderConfig struct {
	Issuer                   string
	AuthorizationEndpoint    string
	TokenEndpoint            string
	RevocationEndpoint       string
	IntrospectionEndpoint    string
	JWKSURI                  string
	ScopesSupported          []string
	GrantTypesSupported      []string
	TokenEndpointAuthMethods []string
	InsecureAllowHTTP        bool
}

func validateProvider(provider *Provider) error {
	if provider == nil {
		return fmt.Errorf("oauth2 provider is required")
	}

	if provider.TokenEndpoint == "" {
		return fmt.Errorf("oauth2 provider token endpoint is required")
	}

	if provider.Issuer != "" {
		if _, err := parseIssuerURL(provider.Issuer, provider.insecureAllowHTTP); err != nil {
			return err
		}
	}

	endpoints := []struct {
		name  string
		value string
	}{
		{name: "authorization endpoint", value: provider.AuthorizationEndpoint},
		{name: "revocation endpoint", value: provider.RevocationEndpoint},
		{name: "introspection endpoint", value: provider.IntrospectionEndpoint},
		{name: "JWKS URI", value: provider.JWKSURI},
	}

	for _, endpoint := range endpoints {
		if endpoint.value == "" {
			continue
		}

		if _, err := parseProviderURL(endpoint.name, endpoint.value, provider.insecureAllowHTTP); err != nil {
			return err
		}
	}

	if _, err := parseTokenEndpointURL(provider.TokenEndpoint, provider.insecureAllowHTTP); err != nil {
		return err
	}

	return nil
}

func cloneStrings(values []string) []string {
	if values == nil {
		return nil
	}

	cloned := make([]string, len(values))
	copy(cloned, values)

	return cloned
}
