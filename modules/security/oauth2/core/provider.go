package core

import (
	"fmt"
	"io"

	"github.com/vmihailenco/msgpack/v5"
)

type (
	// Provider describes an OAuth authorization server and its advertised
	// capabilities.
	Provider struct {
		Issuer                   string
		AuthorizationEndpoint    string
		TokenEndpoint            string
		RevocationEndpoint       string
		IntrospectionEndpoint    string
		JWKSURI                  string
		ScopesSupported          []string
		GrantTypesSupported      []string
		TokenEndpointAuthMethods []string

		insecureAllowHTTP bool
	}

	safeProvider struct {
		Type                     string   `json:"type" msgpack:"type"`
		Issuer                   string   `json:"issuer,omitempty" msgpack:"issuer,omitempty"`
		AuthorizationEndpoint    string   `json:"authorization_endpoint,omitempty" msgpack:"authorization_endpoint,omitempty"`
		TokenEndpoint            string   `json:"token_endpoint" msgpack:"token_endpoint"`
		RevocationEndpoint       string   `json:"revocation_endpoint,omitempty" msgpack:"revocation_endpoint,omitempty"`
		IntrospectionEndpoint    string   `json:"introspection_endpoint,omitempty" msgpack:"introspection_endpoint,omitempty"`
		JWKSURI                  string   `json:"jwks_uri,omitempty" msgpack:"jwks_uri,omitempty"`
		ScopesSupported          []string `json:"scopes_supported,omitempty" msgpack:"scopes_supported,omitempty"`
		GrantTypesSupported      []string `json:"grant_types_supported,omitempty" msgpack:"grant_types_supported,omitempty"`
		TokenEndpointAuthMethods []string `json:"token_endpoint_auth_methods_supported,omitempty" msgpack:"token_endpoint_auth_methods_supported,omitempty"`
	}
)

// NewProvider constructs and validates an OAuth provider.
func NewProvider(config ProviderConfig) (*Provider, error) {
	provider := &Provider{
		Issuer:                   config.Issuer,
		AuthorizationEndpoint:    config.AuthorizationEndpoint,
		TokenEndpoint:            config.TokenEndpoint,
		RevocationEndpoint:       config.RevocationEndpoint,
		IntrospectionEndpoint:    config.IntrospectionEndpoint,
		JWKSURI:                  config.JWKSURI,
		ScopesSupported:          cloneStrings(config.ScopesSupported),
		GrantTypesSupported:      cloneStrings(config.GrantTypesSupported),
		TokenEndpointAuthMethods: cloneStrings(config.TokenEndpointAuthMethods),
		insecureAllowHTTP:        config.InsecureAllowHTTP,
	}

	if err := provider.Validate(); err != nil {
		return nil, err
	}

	return provider, nil
}

// Validate revalidates the provider's exported, mutable configuration.
func (p *Provider) Validate() error {
	return validateProvider(p)
}

// Clone returns a defensive copy of the provider.
func (p *Provider) Clone() *Provider {
	if p == nil {
		return nil
	}

	return &Provider{
		Issuer:                   p.Issuer,
		AuthorizationEndpoint:    p.AuthorizationEndpoint,
		TokenEndpoint:            p.TokenEndpoint,
		RevocationEndpoint:       p.RevocationEndpoint,
		IntrospectionEndpoint:    p.IntrospectionEndpoint,
		JWKSURI:                  p.JWKSURI,
		ScopesSupported:          cloneStrings(p.ScopesSupported),
		GrantTypesSupported:      cloneStrings(p.GrantTypesSupported),
		TokenEndpointAuthMethods: cloneStrings(p.TokenEndpointAuthMethods),
		insecureAllowHTTP:        p.insecureAllowHTTP,
	}
}

// String returns a secret-safe representation of the provider.
func (p *Provider) String() string {
	data, err := p.MarshalJSON()
	if err != nil {
		return `{"type":"oauth2.Provider"}`
	}

	return string(data)
}

// Format keeps all fmt formatting variants on the secret-safe representation.
func (p *Provider) Format(state fmt.State, _ rune) {
	_, _ = io.WriteString(state, p.String())
}

// MarshalJSON returns the provider's safe public configuration.
func (p *Provider) MarshalJSON() ([]byte, error) {
	if p == nil {
		return []byte("null"), nil
	}

	return marshalSafeJSON(p.safeProjection())
}

// MarshalMsgpack serializes the provider's safe public configuration.
func (p *Provider) MarshalMsgpack() ([]byte, error) {
	if p == nil {
		return msgpack.Marshal(nil)
	}

	return msgpack.Marshal(p.safeProjection())
}

func (p *Provider) safeProjection() safeProvider {
	return safeProvider{
		Type:                     "oauth2.Provider",
		Issuer:                   p.Issuer,
		AuthorizationEndpoint:    p.AuthorizationEndpoint,
		TokenEndpoint:            p.TokenEndpoint,
		RevocationEndpoint:       p.RevocationEndpoint,
		IntrospectionEndpoint:    p.IntrospectionEndpoint,
		JWKSURI:                  p.JWKSURI,
		ScopesSupported:          p.ScopesSupported,
		GrantTypesSupported:      p.GrantTypesSupported,
		TokenEndpointAuthMethods: p.TokenEndpointAuthMethods,
	}
}
