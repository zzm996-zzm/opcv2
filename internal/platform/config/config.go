package config

import (
	"errors"
	"os"
	"strings"
)

type Config struct {
	Environment    string
	HTTPAddr       string
	DatabaseURL    string
	RedisAddr      string
	JWTSecret      string
	SMSProvider    string
	SMSDevCode     string
	AutoMigrate    bool
	MigrationsPath string
}

func Load() (Config, error) {
	environment := envOrDefault("OPCV2_ENV", "development")
	smsProvider := envOrDefault("OPCV2_SMS_PROVIDER", "development")
	if environment == "production" && os.Getenv("OPCV2_SMS_PROVIDER") == "" {
		smsProvider = "disabled"
	}
	cfg := Config{
		Environment:    environment,
		HTTPAddr:       envOrDefault("OPCV2_HTTP_ADDR", ":8080"),
		DatabaseURL:    envOrDefault("OPCV2_DATABASE_URL", "postgres://opcv2:opcv2@localhost:5432/opcv2?sslmode=disable"),
		RedisAddr:      envOrDefault("OPCV2_REDIS_ADDR", "localhost:6379"),
		JWTSecret:      envOrDefault("OPCV2_JWT_SECRET", "development-only-change-me"),
		SMSProvider:    smsProvider,
		SMSDevCode:     envOrDefault("OPCV2_SMS_DEV_CODE", "246810"),
		AutoMigrate:    envBoolOrDefault("OPCV2_AUTO_MIGRATE", true),
		MigrationsPath: envOrDefault("OPCV2_MIGRATIONS_PATH", "migrations"),
	}
	if cfg.Environment == "production" {
		if os.Getenv("OPCV2_JWT_SECRET") == "" {
			return Config{}, errors.New("OPCV2_JWT_SECRET is required in production")
		}
		if cfg.SMSProvider == "development" {
			return Config{}, errors.New("OPCV2_SMS_PROVIDER=development is not allowed in production")
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

func envBoolOrDefault(name string, fallback bool) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(name)))
	if value == "" {
		return fallback
	}
	return value == "1" || value == "true" || value == "yes" || value == "on"
}
