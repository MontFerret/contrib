package lib

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/MontFerret/contrib/modules/security/oauth2/core"
	"github.com/MontFerret/contrib/pkg/common/object"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

type (
	providerInput struct {
		Issuer                            string   `json:"issuer"`
		AuthorizationEndpoint             string   `json:"authorizationEndpoint"`
		TokenEndpoint                     string   `json:"tokenEndpoint"`
		RevocationEndpoint                string   `json:"revocationEndpoint"`
		IntrospectionEndpoint             string   `json:"introspectionEndpoint"`
		JWKSURI                           string   `json:"jwksURI"`
		ScopesSupported                   []string `json:"scopesSupported"`
		GrantTypesSupported               []string `json:"grantTypesSupported"`
		TokenEndpointAuthMethodsSupported []string `json:"tokenEndpointAuthMethodsSupported"`
		InsecureAllowHTTP                 bool     `json:"insecureAllowHTTP"`
	}

	clientInput struct {
		ClientID     string `json:"clientID"`
		ClientSecret string `json:"clientSecret"`
		AuthMethod   string `json:"authMethod"`
	}

	discoveryOptionsInput struct {
		Timeout           runtime.Value `json:"timeout"`
		InsecureAllowHTTP bool          `json:"insecureAllowHTTP"`
	}

	clientCredentialsOptionsInput struct {
		Scope      runtime.Value `json:"scope"`
		Parameters runtime.Value `json:"parameters"`
		Timeout    runtime.Value `json:"timeout"`
		Audience   string        `json:"audience"`
	}

	refreshOptionsInput struct {
		Scope        runtime.Value `json:"scope"`
		Parameters   runtime.Value `json:"parameters"`
		Timeout      runtime.Value `json:"timeout"`
		RefreshToken string        `json:"refreshToken"`
	}

	expiredOptionsInput struct {
		Skew runtime.Value `json:"skew"`
	}
)

func decodeProviderConfig(ctx context.Context, value runtime.Value) (core.ProviderConfig, error) {
	var input providerInput
	if err := decodeObject(ctx, value, &input); err != nil {
		return core.ProviderConfig{}, err
	}

	return core.ProviderConfig{
		Issuer:                   input.Issuer,
		AuthorizationEndpoint:    input.AuthorizationEndpoint,
		TokenEndpoint:            input.TokenEndpoint,
		RevocationEndpoint:       input.RevocationEndpoint,
		IntrospectionEndpoint:    input.IntrospectionEndpoint,
		JWKSURI:                  input.JWKSURI,
		ScopesSupported:          input.ScopesSupported,
		GrantTypesSupported:      input.GrantTypesSupported,
		TokenEndpointAuthMethods: input.TokenEndpointAuthMethodsSupported,
		InsecureAllowHTTP:        input.InsecureAllowHTTP,
	}, nil
}

func decodeClientConfig(ctx context.Context, value runtime.Value) (core.ClientConfig, error) {
	var input clientInput
	if err := decodeObject(ctx, value, &input); err != nil {
		return core.ClientConfig{}, err
	}

	return core.ClientConfig{
		ClientID:     input.ClientID,
		ClientSecret: input.ClientSecret,
		AuthMethod:   core.ClientAuthMethod(input.AuthMethod),
	}, nil
}

func decodeDiscoveryOptions(
	ctx context.Context,
	value runtime.Value,
) (core.DiscoveryOptions, error) {
	if value == nil || value == runtime.None {
		return core.DiscoveryOptions{}, nil
	}

	var input discoveryOptionsInput
	if err := decodeObject(ctx, value, &input); err != nil {
		return core.DiscoveryOptions{}, err
	}

	timeout, err := decodeMilliseconds(input.Timeout, "options.timeout")
	if err != nil {
		return core.DiscoveryOptions{}, err
	}

	return core.DiscoveryOptions{
		InsecureAllowHTTP: input.InsecureAllowHTTP,
		Timeout:           timeout,
	}, nil
}

func decodeClientCredentialsOptions(
	ctx context.Context,
	value runtime.Value,
) (core.ClientCredentialsOptions, error) {
	if value == nil || value == runtime.None {
		return core.ClientCredentialsOptions{}, nil
	}

	var input clientCredentialsOptionsInput
	if err := decodeObject(ctx, value, &input); err != nil {
		return core.ClientCredentialsOptions{}, err
	}

	scope, err := decodeScope(ctx, input.Scope, "options.scope")
	if err != nil {
		return core.ClientCredentialsOptions{}, err
	}

	parameters, err := decodeParameters(ctx, input.Parameters, "options.parameters")
	if err != nil {
		return core.ClientCredentialsOptions{}, err
	}

	timeout, err := decodeMilliseconds(input.Timeout, "options.timeout")
	if err != nil {
		return core.ClientCredentialsOptions{}, err
	}

	return core.ClientCredentialsOptions{
		Scope:      scope,
		Audience:   input.Audience,
		Parameters: parameters,
		Timeout:    timeout,
	}, nil
}

func decodeRefreshOptions(
	ctx context.Context,
	value runtime.Value,
) (core.RefreshOptions, error) {
	var input refreshOptionsInput
	if err := decodeObject(ctx, value, &input); err != nil {
		return core.RefreshOptions{}, err
	}

	scope, err := decodeScope(ctx, input.Scope, "options.scope")
	if err != nil {
		return core.RefreshOptions{}, err
	}

	parameters, err := decodeParameters(ctx, input.Parameters, "options.parameters")
	if err != nil {
		return core.RefreshOptions{}, err
	}

	timeout, err := decodeMilliseconds(input.Timeout, "options.timeout")
	if err != nil {
		return core.RefreshOptions{}, err
	}

	return core.RefreshOptions{
		RefreshToken: input.RefreshToken,
		Scope:        scope,
		Parameters:   parameters,
		Timeout:      timeout,
	}, nil
}

func decodeExpiredOptions(
	ctx context.Context,
	value runtime.Value,
) (time.Duration, error) {
	if value == nil || value == runtime.None {
		return 0, nil
	}

	var input expiredOptionsInput
	if err := decodeObject(ctx, value, &input); err != nil {
		return 0, err
	}

	return decodeMilliseconds(input.Skew, "options.skew")
}

func decodeObject(ctx context.Context, value runtime.Value, output any) error {
	obj, err := object.RequireMap(value, "options")
	if err != nil {
		return err
	}

	return sdk.Decode(
		ctx,
		obj,
		output,
		sdk.DisallowUnknownFields(),
		sdk.DisallowNoneValues(),
	)
}

func decodeMilliseconds(value runtime.Value, owner string) (time.Duration, error) {
	if value == nil || value == runtime.None {
		return 0, nil
	}

	milliseconds, ok := value.(runtime.Int)
	if !ok {
		return 0, fmt.Errorf("%s must be an integer number of milliseconds", owner)
	}

	if milliseconds < 0 {
		return 0, fmt.Errorf("%s must be greater than or equal to 0", owner)
	}

	if int64(milliseconds) > math.MaxInt64/int64(time.Millisecond) {
		return 0, fmt.Errorf("%s is too large", owner)
	}

	return time.Duration(milliseconds) * time.Millisecond, nil
}

func decodeScope(ctx context.Context, value runtime.Value, owner string) ([]string, error) {
	if value == nil || value == runtime.None {
		return nil, nil
	}

	if scope, ok := value.(runtime.String); ok {
		return (&core.TokenSet{Scope: scope.String()}).Scopes(), nil
	}

	list, ok := value.(runtime.List)
	if !ok {
		return nil, fmt.Errorf("%s must be a string or an array of strings", owner)
	}

	length, err := list.Length(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", owner, err)
	}

	scope := make([]string, 0, int(length))
	for index := runtime.Int(0); index < length; index++ {
		item, err := list.At(ctx, index)
		if err != nil {
			return nil, fmt.Errorf("%s[%d]: %w", owner, index, err)
		}

		text, ok := item.(runtime.String)
		if !ok {
			return nil, fmt.Errorf("%s[%d] must be a string", owner, index)
		}

		scope = append(scope, text.String())
	}

	return scope, nil
}

func decodeParameters(
	ctx context.Context,
	value runtime.Value,
	owner string,
) (core.Parameters, error) {
	if value == nil || value == runtime.None {
		return nil, nil
	}

	parameters, ok := value.(runtime.Map)
	if !ok {
		return nil, fmt.Errorf("%s must be an object", owner)
	}

	decoded := make(core.Parameters)
	err := parameters.ForEach(ctx, func(
		ctx context.Context,
		value runtime.Value,
		key runtime.Value,
	) (runtime.Boolean, error) {
		name, ok := key.(runtime.String)
		if !ok {
			return false, fmt.Errorf("%s keys must be strings", owner)
		}

		values, err := decodeParameterValue(ctx, value, owner+"."+name.String())
		if err != nil {
			return false, err
		}

		if len(values) > 0 {
			decoded[name.String()] = values
		}

		return true, nil
	})
	if err != nil {
		return nil, err
	}

	return decoded, nil
}

func decodeParameterValue(
	ctx context.Context,
	value runtime.Value,
	owner string,
) ([]string, error) {
	if value == nil || value == runtime.None {
		return nil, nil
	}

	if runtime.IsScalar(value) {
		return []string{value.String()}, nil
	}

	list, ok := value.(runtime.List)
	if !ok {
		return nil, fmt.Errorf("%s must be a scalar or an array of scalars", owner)
	}

	length, err := list.Length(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", owner, err)
	}

	values := make([]string, 0, int(length))
	for index := runtime.Int(0); index < length; index++ {
		item, err := list.At(ctx, index)
		if err != nil {
			return nil, fmt.Errorf("%s[%d]: %w", owner, index, err)
		}

		if item == nil || item == runtime.None {
			continue
		}

		if !runtime.IsScalar(item) {
			return nil, fmt.Errorf("%s[%d] must be a scalar", owner, index)
		}

		values = append(values, item.String())
	}

	return values, nil
}
