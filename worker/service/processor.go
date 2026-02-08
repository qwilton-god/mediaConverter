package service

import (
	"context"
	"path/filepath"
	"time"

	"go.uber.org/zap"

	"mediaConverter/worker/cache"
	"mediaConverter/worker/converter"
	"mediaConverter/worker/kafka"
	"mediaConverter/worker/repository"
)

type Processor struct {
	repo      repository.Repository
	cache     *cache.StatusCache
	logger    *zap.Logger
	converter *converter.Converter
}

func NewProcessor(repo repository.Repository, cache *cache.StatusCache, logger *zap.Logger) *Processor {
	return &Processor{
		repo:      repo,
		cache:     cache,
		logger:    logger,
		converter: converter.NewConverter(logger),
	}
}

func (p *Processor) Process(ctx context.Context, msg *kafka.TaskMessage) error {
	startTime := time.Now()

	if err := p.repo.UpdateTaskStatus(ctx, msg.TaskID, "processing", ""); err != nil {
		return err
	}
	if err := p.cache.Set(ctx, msg.TaskID, "processing"); err != nil {
		return err
	}

	p.logger.Info("Processing task",
		zap.String("task_id", msg.TaskID),
		zap.String("trace_id", msg.TraceID),
		zap.String("file_path", msg.FilePath),
		zap.String("output_format", msg.OutputFormat),
	)

	inputPath := msg.FilePath

	ext := filepath.Ext(inputPath)
	if msg.OutputFormat != "" {
		ext = "." + msg.OutputFormat
	}
	outputPath := "/uploads/" + msg.TaskID + ext

	if err := p.converter.Convert(inputPath, outputPath, msg.OutputFormat, msg.TargetWidth, msg.TargetHeight, msg.Crop); err != nil {
		p.logger.Error("Failed to convert image",
			zap.String("task_id", msg.TaskID),
			zap.Error(err),
		)
		if err := p.repo.UpdateTaskStatus(ctx, msg.TaskID, "failed", err.Error()); err != nil {
			return err
		}
		if err := p.cache.Set(ctx, msg.TaskID, "failed"); err != nil {
			return err
		}
		return err
	}

	if err := p.repo.UpdateTaskStatus(ctx, msg.TaskID, "completed", ""); err != nil {
		return err
	}
	if err := p.cache.Set(ctx, msg.TaskID, "completed"); err != nil {
		return err
	}

	duration := time.Since(startTime)
	p.logger.Info("Task completed",
		zap.String("task_id", msg.TaskID),
		zap.Duration("duration", duration),
	)

	return nil
}
