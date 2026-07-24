package core

import (
	"fmt"
	"slices"
)

// ClientConfig configures an OAuth client.
type ClientConfig struct {
	ClientID     string
	ClientSecret string
	AuthMethod   ClientAuthMethod
}

func normalizeClientConfig(config ClientConfig) ClientConfig {
	if config.AuthMethod != "" {
		return config
	}

	if config.ClientSecret == "" {
		config.AuthMethod = ClientAuthMethodNone
	} else {
		config.AuthMethod = ClientAuthMethodBasic
	}

	return config
}

func validateClient(client *Client) error {
	if client == nil {
		return fmt.Errorf("oauth2 client is required")
	}

	if client.Provider == nil {
		return fmt.Errorf("oauth2 client provider is required")
	}

	if err := client.Provider.Validate(); err != nil {
		return fmt.Errorf("oauth2 client provider is invalid: %w", err)
	}

	if client.ClientID == "" {
		return fmt.Errorf("oauth2 client ID is required")
	}

	if !client.AuthMethod.supported() {
		return fmt.Errorf("oauth2 client authentication method %q is unsupported", client.AuthMethod)
	}

	switch client.AuthMethod {
	case ClientAuthMethodBasic, ClientAuthMethodPost:
		if client.ClientSecret == "" {
			return fmt.Errorf("oauth2 client authentication method %q requires a client secret", client.AuthMethod)
		}
	case ClientAuthMethodNone:
		if client.ClientSecret != "" {
			return fmt.Errorf("oauth2 client authentication method %q does not permit a client secret", client.AuthMethod)
		}
	}

	advertised := client.Provider.TokenEndpointAuthMethods
	if advertised != nil && !slices.Contains(advertised, string(client.AuthMethod)) {
		return fmt.Errorf(
			"oauth2 provider does not advertise client authentication method %q",
			client.AuthMethod,
		)
	}

	return nil
}
