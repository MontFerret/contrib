package core

import (
	"context"
	"sync"

	goredis "github.com/redis/go-redis/v9"
)

type fakeClient struct {
	pingFn    func(context.Context) *goredis.StatusCmd
	doFn      func(context.Context, ...any) *goredis.Cmd
	commandFn func(context.Context) *goredis.CommandsInfoCmd
	closeFn   func() error
	mu        sync.Mutex
	closeCall int
}

func (c *fakeClient) Ping(ctx context.Context) *goredis.StatusCmd {
	if c.pingFn != nil {
		return c.pingFn(ctx)
	}

	cmd := goredis.NewStatusCmd(ctx, "ping")
	cmd.SetVal("PONG")

	return cmd
}

func (c *fakeClient) Do(ctx context.Context, args ...any) *goredis.Cmd {
	if c.doFn != nil {
		return c.doFn(ctx, args...)
	}

	cmd := goredis.NewCmd(ctx, args...)
	cmd.SetVal("OK")

	return cmd
}

func (c *fakeClient) Command(ctx context.Context) *goredis.CommandsInfoCmd {
	if c.commandFn != nil {
		return c.commandFn(ctx)
	}

	cmd := goredis.NewCommandsInfoCmd(ctx, "command")
	cmd.SetVal(map[string]*goredis.CommandInfo{})

	return cmd
}

func (c *fakeClient) Close() error {
	c.mu.Lock()
	c.closeCall++
	c.mu.Unlock()

	if c.closeFn != nil {
		return c.closeFn()
	}

	return nil
}

func (c *fakeClient) closeCalls() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.closeCall
}
