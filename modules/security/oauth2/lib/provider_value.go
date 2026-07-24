package lib

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/MontFerret/contrib/modules/security/oauth2/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type providerValue struct {
	target   *core.Provider
	identity uint64
}

func newProviderValue(provider *core.Provider) *providerValue {
	return &providerValue{
		target:   provider.Clone(),
		identity: nextHostValueIdentity(),
	}
}

func (v *providerValue) provider() *core.Provider {
	if v == nil || v.target == nil {
		return nil
	}

	return v.target.Clone()
}

func (v *providerValue) String() string {
	if v == nil || v.target == nil {
		return runtime.None.String()
	}

	return v.target.String()
}

func (v *providerValue) Format(state fmt.State, verb rune) {
	_, _ = fmt.Fprint(state, v.String())
}

func (v *providerValue) Hash() uint64 {
	if v == nil {
		return runtime.None.Hash()
	}

	return v.identity
}

func (v *providerValue) Copy() runtime.Value {
	if v == nil {
		return runtime.None
	}

	return &providerValue{
		target:   v.target,
		identity: v.identity,
	}
}

func (v *providerValue) Get(_ context.Context, key runtime.Value) (runtime.Value, error) {
	if v == nil || v.target == nil {
		return runtime.None, nil
	}

	switch safeKey(key) {
	case "issuer":
		return stringValue(v.target.Issuer), nil
	case "authorizationEndpoint":
		return stringValue(v.target.AuthorizationEndpoint), nil
	case "tokenEndpoint":
		return stringValue(v.target.TokenEndpoint), nil
	case "revocationEndpoint":
		return stringValue(v.target.RevocationEndpoint), nil
	case "introspectionEndpoint":
		return stringValue(v.target.IntrospectionEndpoint), nil
	case "jwksURI":
		return stringValue(v.target.JWKSURI), nil
	case "scopesSupported":
		return stringArray(v.target.ScopesSupported), nil
	case "grantTypesSupported":
		return stringArray(v.target.GrantTypesSupported), nil
	case "tokenEndpointAuthMethodsSupported":
		return stringArray(v.target.TokenEndpointAuthMethods), nil
	default:
		return runtime.None, nil
	}
}

func (v *providerValue) Unwrap() any {
	if v == nil || v.target == nil {
		return nil
	}

	return v.safeProjection()
}

func (v *providerValue) MarshalJSON() ([]byte, error) {
	if v == nil || v.target == nil {
		return []byte("null"), nil
	}

	return json.Marshal(v.safeProjection())
}

func (v *providerValue) safeProjection() map[string]any {
	return map[string]any{
		"type":                                  "oauth2.Provider",
		"issuer":                                v.target.Issuer,
		"authorization_endpoint":                v.target.AuthorizationEndpoint,
		"token_endpoint":                        v.target.TokenEndpoint,
		"revocation_endpoint":                   v.target.RevocationEndpoint,
		"introspection_endpoint":                v.target.IntrospectionEndpoint,
		"jwks_uri":                              v.target.JWKSURI,
		"scopes_supported":                      append([]string(nil), v.target.ScopesSupported...),
		"grant_types_supported":                 append([]string(nil), v.target.GrantTypesSupported...),
		"token_endpoint_auth_methods_supported": append([]string(nil), v.target.TokenEndpointAuthMethods...),
	}
}
