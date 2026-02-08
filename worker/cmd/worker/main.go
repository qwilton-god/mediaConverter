package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mediaConverter/worker/cache"
	"mediaConverter/worker/kafka"
	"mediaConverter/worker/repository"
	"mediaConverter/worker/service"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	brokers := []string{"kafka:9092"}
	topic := "media_tasks"
	groupID := "worker-group"

	dbURL := "postgres://user:password@postgres:5432/mediadb?sslmode=disable"
	redisAddr := "redis:6379"

	connectCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := pgxpool.New(connectCtx, dbURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	defer redisClient.Close()

	consumer, err := kafka.NewConsumer(brokers, groupID)
	if err != nil {
		logger.Fatal("Failed to create consumer", zap.Error(err))
	}
	defer consumer.Close()

	repo := repository.NewPostgresRepo(db)
	statusCache := cache.NewStatusCache(redisClient)
	processor := service.NewProcessor(repo, statusCache)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler := func(ctx context.Context, msg *kafka.TaskMessage) error {
		logger.Info("Processing task",
			zap.String("task_id", msg.TaskID),
			zap.String("trace_id", msg.TraceID),
		)
		if err := processor.Process(ctx, msg); err != nil {
			logger.Error("Failed to process task",
				zap.String("task_id", msg.TaskID),
				zap.String("trace_id", msg.TraceID),
				zap.Error(err),
			)
		} else {
			logger.Info("Task completed",
				zap.String("task_id", msg.TaskID),
			)
		}
		return nil
	}

	go func() {
		logger.Info("Worker started", zap.String("topic", topic))
		if err := consumer.Consume(ctx, topic, handler); err != nil {
			logger.Error("Consumer error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down worker...")
	cancel()
}
