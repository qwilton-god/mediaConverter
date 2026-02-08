package main

import (
	"context"
	"embed"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mediaConverter/api/cache"
	"mediaConverter/api/config"
	"mediaConverter/api/database"
	"mediaConverter/api/handlers"
	"mediaConverter/api/kafka"
	"mediaConverter/api/middleware"
	"mediaConverter/api/repository"
	"mediaConverter/api/service"

	"go.uber.org/zap"
)

//go:embed static
var staticFS embed.FS

func main() {
	cfg := config.Load()
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logger.Info("API Service starting",
		zap.String("port", cfg.Port),
		zap.String("kafka_brokers", cfg.KafkaBrokers),
	)

	db, err := database.Connect(context.Background(), cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	redisCache, err := database.ConnectCache(cfg.RedisAddr)
	if err != nil {
		logger.Fatal("Failed to connect to cache", zap.Error(err))
	}
	defer redisCache.Close()

	repo := repository.NewPostgresRepo(db)
	statusCache := cache.NewStatusCache(redisCache)
	kafkaProducer, err := kafka.NewProducer([]string{cfg.KafkaBrokers})
	if err != nil {
		logger.Fatal("Failed to connect to Kafka", zap.Error(err))
	}
	defer kafkaProducer.Close()

	taskService := service.NewTaskService(repo, statusCache, kafkaProducer)
	taskHandler := handlers.NewTaskHandler(taskService, logger)

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		content, err := staticFS.ReadFile("static/index.html")
		if err != nil {
			logger.Error("Failed to read static file", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write(content)
	})

	mux.HandleFunc("/upload", taskHandler.Upload)
	mux.HandleFunc("/status/", taskHandler.Status)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	var handler http.Handler = mux
	handler = middleware.TraceID(handler)
	handler = middleware.Logging(logger)(handler)
	handler = middleware.Recovery(logger)(handler)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: handler,
	}

	go func() {
		logger.Info("Server started", zap.String("address", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}
