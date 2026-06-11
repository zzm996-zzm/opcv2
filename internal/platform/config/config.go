package config

import (
	"errors"
	"os"
)

type Config struct {
	Environment string
	HTTPAddr    string
	DatabaseURL string
	RedisAddr   string
	JWTSecret   string
	SMSProvider string
	SMSDevCode  string
}

func Load() (Config, error) {
	cfg := Config{
		Environment: envOrDefault("OPCV2_ENV", "development"),
		HTTPAddr:    envOrDefault("OPCV2_HTTP_ADDR", ":8080"),
		DatabaseURL: envOrDefault("OPCV2_DATABASE_URL", "postgres://opcv2:opcv2@localhost:5432/opcv2?sslmode=disable"),
		RedisAddr:   envOrDefault("OPCV2_REDIS_ADDR", "localhost:6379"),
		JWTSecret:   envOrDefault("OPCV2_JWT_SECRET", "development-only-change-me"),
		SMSProvider: envOrDefault("OPCV2_SMS_PROVIDER", "development"),
		SMSDevCode:  envOrDefault("OPCV2_SMS_DEV_CODE", "246810"),
	}
	if cfg.Environment == "production" {
		if os.Getenv("OPCV2_JWT_SECRET") == "" {
			return Config{}, errors.New("OPCV2_JWT_SECRET is required in production")
		}
		if os.Getenv("OPCV2_SMS_PROVIDER") == "" || cfg.SMSProvider == "development" {
			return Config{}, errors.New("a production SMS provider is required")
		}
	}
	return cfg, nil
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
