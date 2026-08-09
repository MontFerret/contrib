package core

import (
	"context"
	"fmt"
	"strings"
	"sync"

	goredis "github.com/redis/go-redis/v9"

	commonresource "github.com/MontFerret/contrib/pkg/common/resource"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Connection is an opaque Redis connection handle exposed to Ferret.
type Connection struct {
	client        redisClient
	commandInfo   map[string]*goredis.CommandInfo
	mu            sync.RWMutex
	commandInfoMu sync.Mutex
	id            uint64
	closed        bool
}

var (
	_ runtime.Value     = (*Connection)(nil)
	_ runtime.Resource  = (*Connection)(nil)
	_ runtime.Queryable = (*Connection)(nil)
)

func newConnection(client redisClient) *Connection {
	return &Connection{
		client: client,
		id:     newResourceID(),
	}
}

func (c *Connection) Query(ctx context.Context, q runtime.Query) (runtime.List, error) {
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

func (c *Connection) QueryOne(ctx context.Context, q runtime.Query) (runtime.Value, error) {
	return runtime.DefaultQueryOne(ctx, q, c.Query)
}

func (c *Connection) QueryCount(ctx context.Context, q runtime.Query) (runtime.Int, error) {
	return runtime.DefaultQueryCount(ctx, q, c.Query)
}

func (c *Connection) QueryExists(ctx context.Context, q runtime.Query) (runtime.Boolean, error) {
	return runtime.DefaultQueryExists(ctx, q, c.Query)
}

func (c *Connection) Close() error {
	c.mu.Lock()

	if c.closed {
		c.mu.Unlock()

		return nil
	}

	c.closed = true
	client := c.client
	c.mu.Unlock()

	if err := client.Close(); err != nil {
		return OperationError("CLOSE", err)
	}

	return nil
}

func (c *Connection) ResourceID() uint64 {
	return c.id
}

func (c *Connection) String() string {
	return commonresource.Display("db.redis.connection")
}

func (c *Connection) Hash() uint64 {
	return commonresource.Hash("db.redis.connection", c.id)
}

func (c *Connection) Copy() runtime.Value {
	return c
}

func (c *Connection) MarshalJSON() ([]byte, error) {
	return commonresource.MarshalDisplayJSON("db.redis.connection")
}

func (c *Connection) redisClient(operation string) (redisClient, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, OperationError(operation, errConnectionClosed)
	}

	return c.client, nil
}

func (c *Connection) readCommandInfo(
	ctx context.Context,
	client redisClient,
	command string,
) (*goredis.CommandInfo, error) {
	c.commandInfoMu.Lock()
	defer c.commandInfoMu.Unlock()

	name := strings.ToLower(command)
	if c.commandInfo == nil {
		catalog, err := loadCommandInfo(ctx, client)
		if err != nil {
			return nil, fmt.Errorf(
				"cannot validate command %q as read-only; use redis_exec: %w",
				command,
				err,
			)
		}

		c.commandInfo = catalog
	}

	if info := c.commandInfo[name]; info != nil {
		return info, nil
	}

	// Refresh once for a missing command so commands loaded by Redis modules
	// after this connection was opened can participate in read validation.
	catalog, err := loadCommandInfo(ctx, client)
	if err != nil {
		return nil, fmt.Errorf(
			"cannot refresh read-only metadata for command %q; use redis_exec: %w",
			command,
			err,
		)
	}

	c.commandInfo = catalog
	if info := c.commandInfo[name]; info != nil {
		return info, nil
	}

	return nil, fmt.Errorf("command %q is absent from Redis command metadata; use redis_exec", command)
}
