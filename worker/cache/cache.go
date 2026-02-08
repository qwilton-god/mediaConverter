package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type StatusCache struct {
	client *redis.Client
}

func NewStatusCache(client *redis.Client) *StatusCache {
	return &StatusCache{client: client}
}

func (c *StatusCache) Set(ctx context.Context, taskID string, status string) error {
	return c.client.Set(ctx, "task:status:"+taskID, status, 0).Err()
}
