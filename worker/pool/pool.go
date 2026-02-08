package pool

import (
	"context"
	"sync"

	"mediaConverter/worker/kafka"
)

type WorkerPool struct {
	sem chan struct{}
	wg  sync.WaitGroup
}

func NewWorkerPool(maxWorkers int) *WorkerPool {
	return &WorkerPool{
		sem: make(chan struct{}, maxWorkers),
	}
}

func (p *WorkerPool) Submit(ctx context.Context, msg *kafka.TaskMessage, handler func(context.Context, *kafka.TaskMessage) error) {
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()

		select {
		case p.sem <- struct{}{}:
			defer func() { <-p.sem }()
			handler(ctx, msg)
		case <-ctx.Done():
		}
	}()
}

func (p *WorkerPool) Wait() {
	p.wg.Wait()
}
