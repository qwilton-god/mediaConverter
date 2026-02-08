package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"mediaConverter/api/database"
	"mediaConverter/api/models"
)

const (
	statusKeyPrefix = "task:status:"
	statusTTL       = 10 * time.Minute
)

type StatusCache struct {
	cache *database.Cache
}

func NewStatusCache(cache *database.Cache) *StatusCache {
	return &StatusCache{cache: cache}
}

func (sc *StatusCache) Get(ctx context.Context, taskID string) (*models.TaskStatus, error) {
	key := fmt.Sprintf("%s%s", statusKeyPrefix, taskID)

	data, err := sc.cache.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var status models.TaskStatus
	if err := json.Unmarshal([]byte(data), &status); err != nil {
		status = models.TaskStatus(data)
	}

	return &status, nil
}

func (sc *StatusCache) Set(ctx context.Context, taskID string, status models.TaskStatus) error {
	key := fmt.Sprintf("%s%s", statusKeyPrefix, taskID)

	data, err := json.Marshal(status)
	if err != nil {
		return err
	}

	return sc.cache.Set(ctx, key, data, statusTTL)
}

func (sc *StatusCache) Delete(ctx context.Context, taskID string) error {
	key := fmt.Sprintf("%s%s", statusKeyPrefix, taskID)
	return sc.cache.Del(ctx, key)
}
