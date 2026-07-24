package core

import (
	"fmt"
	"io"

	"github.com/vmihailenco/msgpack/v5"
)

type (
	// Client contains an OAuth provider and the credentials used at its token
	// endpoint.
	Client struct {
		Provider     *Provider
		ClientID     string
		ClientSecret string
		AuthMethod   ClientAuthMethod
	}

	safeClient struct {
		Type         string           `json:"type" msgpack:"type"`
		Provider     *Provider        `json:"provider,omitempty" msgpack:"provider,omitempty"`
		ClientID     string           `json:"client_id" msgpack:"client_id"`
		ClientSecret string           `json:"client_secret,omitempty" msgpack:"client_secret,omitempty"`
		AuthMethod   ClientAuthMethod `json:"auth_method" msgpack:"auth_method"`
	}
)

// NewClient constructs and validates an OAuth client.
func NewClient(provider *Provider, config ClientConfig) (*Client, error) {
	config = normalizeClientConfig(config)

	client := &Client{
		Provider:     provider.Clone(),
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		AuthMethod:   config.AuthMethod,
	}

	if err := client.Validate(); err != nil {
		return nil, err
	}

	return client, nil
}

// Validate revalidates the client's exported, mutable configuration.
func (c *Client) Validate() error {
	return validateClient(c)
}

// Clone returns a defensive copy of the client and its provider.
func (c *Client) Clone() *Client {
	if c == nil {
		return nil
	}

	return &Client{
		Provider:     c.Provider.Clone(),
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		AuthMethod:   c.AuthMethod,
	}
}

// String returns a representation that redacts the client secret.
func (c *Client) String() string {
	data, err := c.MarshalJSON()
	if err != nil {
		return `{"type":"oauth2.Client"}`
	}

	return string(data)
}

// Format keeps all fmt formatting variants on the secret-safe representation.
func (c *Client) Format(state fmt.State, _ rune) {
	_, _ = io.WriteString(state, c.String())
}

// MarshalJSON returns a safe representation of the client.
func (c *Client) MarshalJSON() ([]byte, error) {
	if c == nil {
		return []byte("null"), nil
	}

	return marshalSafeJSON(c.safeProjection())
}

// MarshalMsgpack serializes a redacted client view.
func (c *Client) MarshalMsgpack() ([]byte, error) {
	if c == nil {
		return msgpack.Marshal(nil)
	}

	return msgpack.Marshal(c.safeProjection())
}

func (c *Client) safeProjection() safeClient {
	secret := ""
	if c.ClientSecret != "" {
		secret = "<redacted>"
	}

	return safeClient{
		Type:         "oauth2.Client",
		Provider:     c.Provider,
		ClientID:     c.ClientID,
		ClientSecret: secret,
		AuthMethod:   c.AuthMethod,
	}
}
