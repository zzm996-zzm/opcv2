package config

import (
	"errors"
	"os"
	"strings"
)

type Config struct {
	Environment              string
	HTTPAddr                 string
	DatabaseURL              string
	RedisAddr                string
	JWTSecret                string
	SMSProvider              string
	SMSDevCode               string
	AIProvider               string
	AIModel                  string
	AIAPIKey                 string
	AIBaseURL                string
	AITimeoutSeconds         int
	LeadProvider             string
	TianyanchaAPIKey         string
	TianyanchaBaseURL        string
	TianyanchaTimeoutSeconds int
	SerperAPIKey             string
	SerperBaseURL            string
	SerperTimeoutSeconds     int
	AutoMigrate              bool
	MigrationsPath           string
	CORSAllowedOrigins       []string
	ExpensiveEndpointLimit   int
}

func Load() (Config, error) {
	environment := envOrDefault("OPCV2_ENV", "development")
	smsProvider := envOrDefault("OPCV2_SMS_PROVIDER", "development")
	if environment == "production" && os.Getenv("OPCV2_SMS_PROVIDER") == "" {
		smsProvider = "disabled"
	}
	cfg := Config{
		Environment:              environment,
		HTTPAddr:                 envOrDefault("OPCV2_HTTP_ADDR", ":8080"),
		DatabaseURL:              envOrDefault("OPCV2_DATABASE_URL", "postgres://opcv2:opcv2@localhost:5432/opcv2?sslmode=disable"),
		RedisAddr:                envOrDefault("OPCV2_REDIS_ADDR", "localhost:6379"),
		JWTSecret:                envOrDefault("OPCV2_JWT_SECRET", "development-only-change-me"),
		SMSProvider:              smsProvider,
		SMSDevCode:               envOrDefault("OPCV2_SMS_DEV_CODE", "246810"),
		AIProvider:               envOrDefault("OPCV2_AI_PROVIDER", "development"),
		AIModel:                  envOrDefault("OPCV2_AI_MODEL", "development-model"),
		AIAPIKey:                 os.Getenv("OPCV2_AI_API_KEY"),
		AIBaseURL:                os.Getenv("OPCV2_AI_BASE_URL"),
		AITimeoutSeconds:         envIntOrDefault("OPCV2_AI_TIMEOUT_SECONDS", 30),
		LeadProvider:             envOrDefault("OPCV2_LEAD_PROVIDER", "development"),
		TianyanchaAPIKey:         os.Getenv("OPCV2_TYC_API_KEY"),
		TianyanchaBaseURL:        os.Getenv("OPCV2_TYC_BASE_URL"),
		TianyanchaTimeoutSeconds: envIntOrDefault("OPCV2_TYC_TIMEOUT_SECONDS", 10),
		SerperAPIKey:             os.Getenv("OPCV2_SERPER_API_KEY"),
		SerperBaseURL:            os.Getenv("OPCV2_SERPER_BASE_URL"),
		SerperTimeoutSeconds:     envIntOrDefault("OPCV2_SERPER_TIMEOUT_SECONDS", 10),
		AutoMigrate:              envBoolOrDefault("OPCV2_AUTO_MIGRATE", true),
		MigrationsPath:           envOrDefault("OPCV2_MIGRATIONS_PATH", "migrations"),
		CORSAllowedOrigins:       envCSVOrDefault("OPCV2_CORS_ALLOWED_ORIGINS", []string{"http://localhost:5173"}),
		ExpensiveEndpointLimit:   envIntOrDefault("OPCV2_EXPENSIVE_ENDPOINT_LIMIT", 20),
	}
	if cfg.Environment == "production" {
		if os.Getenv("OPCV2_JWT_SECRET") == "" {
			return Config{}, errors.New("OPCV2_JWT_SECRET is required in production")
		}
		if cfg.SMSProvider == "development" {
			return Config{}, errors.New("OPCV2_SMS_PROVIDER=development is not allowed in production")
		}
		if os.Getenv("OPCV2_AI_PROVIDER") == "" {
			return Config{}, errors.New("OPCV2_AI_PROVIDER is required in production")
		}
		if cfg.AIProvider == "development" {
			return Config{}, errors.New("OPCV2_AI_PROVIDER=development is not allowed in production")
		}
		if cfg.AIAPIKey == "" {
			return Config{}, errors.New("OPCV2_AI_API_KEY is required in production")
		}
		if os.Getenv("OPCV2_LEAD_PROVIDER") == "" {
			return Config{}, errors.New("OPCV2_LEAD_PROVIDER is required in production")
		}
		if cfg.LeadProvider == "development" {
			return Config{}, errors.New("OPCV2_LEAD_PROVIDER=development is not allowed in production")
		}
		if cfg.LeadProvider == "tianyancha" && cfg.TianyanchaAPIKey == "" {
			return Config{}, errors.New("OPCV2_TYC_API_KEY is required in production")
		}
		if os.Getenv("OPCV2_CORS_ALLOWED_ORIGINS") == "" {
			return Config{}, errors.New("OPCV2_CORS_ALLOWED_ORIGINS is required in production")
		}
		if hasWildcardOrigin(cfg.CORSAllowedOrigins) {
			return Config{}, errors.New("OPCV2_CORS_ALLOWED_ORIGINS cannot contain wildcard in production")
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

func envIntOrDefault(name string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	var result int
	for _, char := range value {
		if char < '0' || char > '9' {
			return fallback
		}
		result = result*10 + int(char-'0')
	}
	if result <= 0 {
		return fallback
	}
	return result
}

func envCSVOrDefault(name string, fallback []string) []string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	if len(result) == 0 {
		return fallback
	}
	return result
}

func hasWildcardOrigin(origins []string) bool {
	for _, origin := range origins {
		if origin == "*" {
			return true
		}
	}
	return false
}
