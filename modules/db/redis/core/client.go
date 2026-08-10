package core

import (
	"context"

	goredis "github.com/redis/go-redis/v9"
)

type redisClient interface {
	Ping(ctx context.Context) *goredis.StatusCmd
	Do(ctx context.Context, args ...any) *goredis.Cmd
	Command(ctx context.Context) *goredis.CommandsInfoCmd
	Close() error
}
