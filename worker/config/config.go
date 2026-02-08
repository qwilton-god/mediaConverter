package config

import (
	"os"
	"strconv"
)

type Config struct {
	KafkaBrokers string
	KafkaTopic   string
	KafkaGroupID string
	DatabaseURL  string
	RedisAddr    string
	WorkerCount  int
}

func Load() *Config {
	return &Config{
		KafkaBrokers: getEnv("KAFKA_BROKERS", "localhost:9092"),
		KafkaTopic:   getEnv("KAFKA_TOPIC", "media_tasks"),
		KafkaGroupID: getEnv("KAFKA_GROUP_ID", "worker-group"),
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/mediadb?sslmode=disable"),
		RedisAddr:    getEnv("REDIS_ADDR", "localhost:6379"),
		WorkerCount:  getEnvAsInt("WORKER_COUNT", 5),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
