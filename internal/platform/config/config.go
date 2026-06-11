package config

import "os"

type Config struct {
	Environment string
	HTTPAddr    string
	DatabaseURL string
	RedisAddr   string
}

func Load() (Config, error) {
	return Config{
		Environment: envOrDefault("OPCV2_ENV", "development"),
		HTTPAddr:    envOrDefault("OPCV2_HTTP_ADDR", ":8080"),
		DatabaseURL: envOrDefault("OPCV2_DATABASE_URL", "postgres://opcv2:opcv2@localhost:5432/opcv2?sslmode=disable"),
		RedisAddr:   envOrDefault("OPCV2_REDIS_ADDR", "localhost:6379"),
	}, nil
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
