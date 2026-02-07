package main

import (
	"time"

	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logger.Info("Worker Service starting...")

	for {
		logger.Info("Worker is running...")
		time.Sleep(10 * time.Second)
	}
}
