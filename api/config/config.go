package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port         string
	Env          string
	KafkaBrokers string
	DatabaseURL  string
	RedisAddr    string
	MaxFileSize  int64
}

func Load() *Config {
	return &Config{
		Port:         getEnv("SERVICE_PORT", "8081"),
		Env:          getEnv("ENV", "development"),
		KafkaBrokers: getEnv("KAFKA_BROKERS", "localhost:9092"),
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/mediadb?sslmode=disable"),
		RedisAddr:    getEnv("REDIS_ADDR", "localhost:6379"),
		MaxFileSize:  getEnvAsInt64("MAX_FILE_SIZE", 100*1024*1024),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intVal
		}
	}
	return defaultValue
}
