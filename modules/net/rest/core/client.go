package core

import (
	"context"
	"encoding/binary"
	"hash/fnv"
	"net/http"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Client is an opaque HTTP API client exposed to Ferret.
type Client struct {
	httpClient *http.Client
	config     Config
	id         uint64
}

// NewClient creates a configured HTTP API client handle.
func NewClient(config Config) *Client {
	return &Client{
		config:     config,
		httpClient: http.DefaultClient,
		id:         newResourceID(),
	}
}

func (c *Client) Query(ctx context.Context, q runtime.Query) (runtime.List, error) {
	value, flatten, err := executeQuery(ctx, c, q)
	if err != nil {
		return nil, err
	}

	if flatten {
		if list, ok := value.(runtime.List); ok {
			return list, nil
		}
	}

	return runtime.NewArrayWith(value), nil
}

func (c *Client) QueryOne(ctx context.Context, q runtime.Query) (runtime.Value, error) {
	return runtime.DefaultQueryOne(ctx, q, c.Query)
}

func (c *Client) QueryCount(ctx context.Context, q runtime.Query) (runtime.Int, error) {
	return runtime.DefaultQueryCount(ctx, q, c.Query)
}

func (c *Client) QueryExists(ctx context.Context, q runtime.Query) (runtime.Boolean, error) {
	return runtime.DefaultQueryExists(ctx, q, c.Query)
}

func (c *Client) ResourceID() uint64 {
	return c.id
}

func (c *Client) String() string {
	return "<http.client>"
}

func (c *Client) Hash() uint64 {
	h := fnv.New64a()
	h.Write([]byte("http.client:"))

	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, c.id)
	h.Write(bytes)

	return h.Sum64()
}

func (c *Client) Copy() runtime.Value {
	return c
}

func (c *Client) MarshalJSON() ([]byte, error) {
	return []byte(`"<http.client>"`), nil
}
