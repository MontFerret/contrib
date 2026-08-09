package core

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	goredis "github.com/redis/go-redis/v9"
)

type clientFactory func(*goredis.Options) redisClient

// Open opens and verifies a Redis connection from validated options.
func Open(ctx context.Context, options OpenOptions) (*Connection, error) {
	return openWithFactory(ctx, options, newRedisClient)
}

func openWithFactory(ctx context.Context, options OpenOptions, factory clientFactory) (*Connection, error) {
	clientOptions, err := parseClientOptions(options.URL)
	if err != nil {
		return nil, OperationError("OPEN", err)
	}

	client := factory(clientOptions)
	if client == nil {
		return nil, OperationError("OPEN", fmt.Errorf("redis client factory returned nil"))
	}

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()

		return nil, OperationError("OPEN", err)
	}

	return newConnection(client), nil
}

func parseClientOptions(input string) (*goredis.Options, error) {
	raw := strings.TrimSpace(input)
	if raw == "" {
		return nil, fmt.Errorf("url is required")
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid Redis URL: %w", err)
	}

	if !strings.EqualFold(parsed.Scheme, "redis") && !strings.EqualFold(parsed.Scheme, "rediss") {
		return nil, fmt.Errorf("unsupported Redis URL scheme %q; expected redis or rediss", parsed.Scheme)
	}

	opts, err := goredis.ParseURL(raw)
	if err != nil {
		return nil, err
	}

	// Ferret Queryable implementations must observe caller cancellation and
	// per-query deadlines while remote work is in progress.
	opts.ContextTimeoutEnabled = true

	return opts, nil
}

func newRedisClient(options *goredis.Options) redisClient {
	return goredis.NewClient(options)
}
