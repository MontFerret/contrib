package core

import (
	"context"
	"strings"

	goredis "github.com/redis/go-redis/v9"
)

func loadCommandInfo(ctx context.Context, client redisClient) (map[string]*goredis.CommandInfo, error) {
	catalog, err := client.Command(ctx).Result()
	if err != nil {
		return nil, err
	}

	normalized := make(map[string]*goredis.CommandInfo, len(catalog))
	for name, info := range catalog {
		normalized[strings.ToLower(name)] = info
	}

	return normalized, nil
}
