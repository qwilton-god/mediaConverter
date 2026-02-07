package main

import (
	"log"
	"net/http"

	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logger.Info("API Service starting", zap.String("port", "8081"))

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("# Metrics\n"))
	})

	logger.Info("Server started", zap.String("address", ":8081"))
	log.Fatal(http.ListenAndServe(":8081", mux))
}
