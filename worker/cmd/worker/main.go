package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mediaConverter/worker/cache"
	"mediaConverter/worker/config"
	"mediaConverter/worker/kafka"
	"mediaConverter/worker/pool"
	"mediaConverter/worker/repository"
	"mediaConverter/worker/service"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg := config.Load()

	connectCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := pgxpool.New(connectCtx, cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})
	defer redisClient.Close()

	consumer, err := kafka.NewConsumer([]string{cfg.KafkaBrokers}, cfg.KafkaGroupID)
	if err != nil {
		logger.Fatal("Failed to create consumer", zap.Error(err))
	}
	defer consumer.Close()

	repo := repository.NewPostgresRepo(db)
	statusCache := cache.NewStatusCache(redisClient)
	processor := service.NewProcessor(repo, statusCache, logger)
	workerPool := pool.NewWorkerPool(cfg.WorkerCount)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler := func(ctx context.Context, msg *kafka.TaskMessage) error {
		return processor.Process(ctx, msg)
	}

	go func() {
		logger.Info("Worker started",
			zap.String("topic", cfg.KafkaTopic),
			zap.Int("worker_count", cfg.WorkerCount),
		)
		if err := consumer.Consume(ctx, cfg.KafkaTopic, handler); err != nil {
			logger.Error("Consumer error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down worker...")
	cancel()
	workerPool.Wait()
	logger.Info("Worker stopped")
}
