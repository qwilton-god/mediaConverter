package service

import (
	"context"
	"time"

	"mediaConverter/worker/cache"
	"mediaConverter/worker/kafka"
	"mediaConverter/worker/repository"
)

type Processor struct {
	repo  repository.Repository
	cache *cache.StatusCache
}

func NewProcessor(repo repository.Repository, cache *cache.StatusCache) *Processor {
	return &Processor{
		repo:  repo,
		cache: cache,
	}
}

func (p *Processor) Process(ctx context.Context, msg *kafka.TaskMessage) error {
	if err := p.repo.UpdateTaskStatus(ctx, msg.TaskID, "processing", ""); err != nil {
		return err
	}
	if err := p.cache.Set(ctx, msg.TaskID, "processing"); err != nil {
		return err
	}

	time.Sleep(3 * time.Second)

	if err := p.repo.UpdateTaskStatus(ctx, msg.TaskID, "completed", ""); err != nil {
		return err
	}
	if err := p.cache.Set(ctx, msg.TaskID, "completed"); err != nil {
		return err
	}

	return nil
}
