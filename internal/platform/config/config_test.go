package config

import "testing"

func TestLoadUsesDevelopmentDefaults(t *testing.T) {
	t.Setenv("OPCV2_ENV", "")
	t.Setenv("OPCV2_HTTP_ADDR", "")
	t.Setenv("OPCV2_DATABASE_URL", "")
	t.Setenv("OPCV2_REDIS_ADDR", "")
	t.Setenv("OPCV2_JWT_SECRET", "")
	t.Setenv("OPCV2_SMS_PROVIDER", "")
	t.Setenv("OPCV2_SMS_DEV_CODE", "")
	t.Setenv("OPCV2_AI_PROVIDER", "")
	t.Setenv("OPCV2_AI_MODEL", "")
	t.Setenv("OPCV2_AI_API_KEY", "")
	t.Setenv("OPCV2_AI_BASE_URL", "")
	t.Setenv("OPCV2_AI_TIMEOUT_SECONDS", "")
	t.Setenv("OPCV2_LEAD_PROVIDER", "")
	t.Setenv("OPCV2_TYC_API_KEY", "")
	t.Setenv("OPCV2_TYC_BASE_URL", "")
	t.Setenv("OPCV2_TYC_TIMEOUT_SECONDS", "")
	t.Setenv("OPCV2_SERPER_API_KEY", "")
	t.Setenv("OPCV2_SERPER_BASE_URL", "")
	t.Setenv("OPCV2_SERPER_TIMEOUT_SECONDS", "")
	t.Setenv("OPCV2_AUTO_MIGRATE", "")
	t.Setenv("OPCV2_MIGRATIONS_PATH", "")
	t.Setenv("OPCV2_CORS_ALLOWED_ORIGINS", "")
	t.Setenv("OPCV2_EXPENSIVE_ENDPOINT_LIMIT", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Environment != "development" {
		t.Fatalf("Environment = %q, want development", cfg.Environment)
	}
	if cfg.HTTPAddr != ":8080" {
		t.Fatalf("HTTPAddr = %q, want :8080", cfg.HTTPAddr)
	}
	if cfg.DatabaseURL != "postgres://opcv2:opcv2@localhost:5432/opcv2?sslmode=disable" {
		t.Fatalf("DatabaseURL = %q, want development database", cfg.DatabaseURL)
	}
	if cfg.RedisAddr != "localhost:6379" {
		t.Fatalf("RedisAddr = %q, want localhost:6379", cfg.RedisAddr)
	}
	if cfg.JWTSecret == "" || cfg.SMSProvider != "development" || cfg.SMSDevCode != "246810" {
		t.Fatalf("development auth defaults = %+v", cfg)
	}
	if !cfg.AutoMigrate || cfg.MigrationsPath != "migrations" {
		t.Fatalf("migration defaults = %+v, want auto migrate using migrations", cfg)
	}
	if cfg.AIProvider != "development" || cfg.AIModel != "development-model" || cfg.AITimeoutSeconds != 30 {
		t.Fatalf("AI defaults = %+v, want development provider/model/timeout", cfg)
	}
	if cfg.LeadProvider != "development" || cfg.TianyanchaTimeoutSeconds != 10 || cfg.SerperTimeoutSeconds != 10 {
		t.Fatalf("lead provider defaults = %+v, want development provider and 10s timeouts", cfg)
	}
	if len(cfg.CORSAllowedOrigins) == 0 || cfg.ExpensiveEndpointLimit != 20 {
		t.Fatalf("security defaults = %+v, want development CORS origins and rate limit", cfg)
	}
}

func TestLoadUsesEnvironmentOverrides(t *testing.T) {
	t.Setenv("OPCV2_ENV", "test")
	t.Setenv("OPCV2_HTTP_ADDR", ":9090")
	t.Setenv("OPCV2_DATABASE_URL", "postgres://example")
	t.Setenv("OPCV2_REDIS_ADDR", "redis:6380")
	t.Setenv("OPCV2_JWT_SECRET", "test-jwt-secret")
	t.Setenv("OPCV2_SMS_PROVIDER", "development")
	t.Setenv("OPCV2_SMS_DEV_CODE", "123456")
	t.Setenv("OPCV2_AI_PROVIDER", "openai")
	t.Setenv("OPCV2_AI_MODEL", "gpt-test")
	t.Setenv("OPCV2_AI_API_KEY", "test-ai-key")
	t.Setenv("OPCV2_AI_BASE_URL", "https://api.example.test")
	t.Setenv("OPCV2_AI_TIMEOUT_SECONDS", "45")
	t.Setenv("OPCV2_LEAD_PROVIDER", "tianyancha")
	t.Setenv("OPCV2_TYC_API_KEY", "test-tyc-key")
	t.Setenv("OPCV2_TYC_BASE_URL", "https://tyc.example.test")
	t.Setenv("OPCV2_TYC_TIMEOUT_SECONDS", "12")
	t.Setenv("OPCV2_SERPER_API_KEY", "test-serper-key")
	t.Setenv("OPCV2_SERPER_BASE_URL", "https://serper.example.test")
	t.Setenv("OPCV2_SERPER_TIMEOUT_SECONDS", "13")
	t.Setenv("OPCV2_AUTO_MIGRATE", "false")
	t.Setenv("OPCV2_MIGRATIONS_PATH", "/app/migrations")
	t.Setenv("OPCV2_CORS_ALLOWED_ORIGINS", "https://app.example.com, https://admin.example.com")
	t.Setenv("OPCV2_EXPENSIVE_ENDPOINT_LIMIT", "7")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Environment != "test" || cfg.HTTPAddr != ":9090" {
		t.Fatalf("environment/http = %q/%q, want test/:9090", cfg.Environment, cfg.HTTPAddr)
	}
	if cfg.DatabaseURL != "postgres://example" || cfg.RedisAddr != "redis:6380" {
		t.Fatalf("database/redis overrides were not loaded: %+v", cfg)
	}
	if cfg.JWTSecret != "test-jwt-secret" || cfg.SMSDevCode != "123456" {
		t.Fatalf("auth overrides were not loaded: %+v", cfg)
	}
	if cfg.AutoMigrate || cfg.MigrationsPath != "/app/migrations" {
		t.Fatalf("migration overrides were not loaded: %+v", cfg)
	}
	if cfg.AIProvider != "openai" || cfg.AIModel != "gpt-test" || cfg.AIAPIKey != "test-ai-key" || cfg.AIBaseURL != "https://api.example.test" || cfg.AITimeoutSeconds != 45 {
		t.Fatalf("AI overrides were not loaded: %+v", cfg)
	}
	if cfg.LeadProvider != "tianyancha" || cfg.TianyanchaAPIKey != "test-tyc-key" || cfg.TianyanchaBaseURL != "https://tyc.example.test" || cfg.TianyanchaTimeoutSeconds != 12 {
		t.Fatalf("Tianyancha overrides were not loaded: %+v", cfg)
	}
	if cfg.SerperAPIKey != "test-serper-key" || cfg.SerperBaseURL != "https://serper.example.test" || cfg.SerperTimeoutSeconds != 13 {
		t.Fatalf("Serper overrides were not loaded: %+v", cfg)
	}
	if len(cfg.CORSAllowedOrigins) != 2 || cfg.CORSAllowedOrigins[0] != "https://app.example.com" || cfg.ExpensiveEndpointLimit != 7 {
		t.Fatalf("security overrides were not loaded: %+v", cfg)
	}
}

func TestLoadRejectsProductionWithoutSecrets(t *testing.T) {
	t.Setenv("OPCV2_ENV", "production")
	t.Setenv("OPCV2_JWT_SECRET", "")
	t.Setenv("OPCV2_SMS_PROVIDER", "")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want production configuration error")
	}
}

func TestLoadDefaultsProductionSMSToDisabled(t *testing.T) {
	t.Setenv("OPCV2_ENV", "production")
	t.Setenv("OPCV2_JWT_SECRET", "production-secret")
	t.Setenv("OPCV2_SMS_PROVIDER", "")
	t.Setenv("OPCV2_AI_PROVIDER", "openai")
	t.Setenv("OPCV2_AI_MODEL", "gpt-production")
	t.Setenv("OPCV2_AI_API_KEY", "production-ai-key")
	t.Setenv("OPCV2_LEAD_PROVIDER", "tianyancha")
	t.Setenv("OPCV2_TYC_API_KEY", "production-tyc-key")
	t.Setenv("OPCV2_CORS_ALLOWED_ORIGINS", "https://app.example.com")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.SMSProvider != "disabled" {
		t.Fatalf("SMSProvider = %q, want disabled", cfg.SMSProvider)
	}
}

func TestLoadRejectsDevelopmentSMSInProduction(t *testing.T) {
	t.Setenv("OPCV2_ENV", "production")
	t.Setenv("OPCV2_JWT_SECRET", "production-secret")
	t.Setenv("OPCV2_SMS_PROVIDER", "development")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want production SMS provider error")
	}
}

func TestLoadRejectsProductionWithoutExplicitCORSOrigins(t *testing.T) {
	t.Setenv("OPCV2_ENV", "production")
	t.Setenv("OPCV2_JWT_SECRET", "production-secret")
	t.Setenv("OPCV2_SMS_PROVIDER", "disabled")
	t.Setenv("OPCV2_AI_PROVIDER", "openai")
	t.Setenv("OPCV2_AI_MODEL", "gpt-production")
	t.Setenv("OPCV2_AI_API_KEY", "production-ai-key")
	t.Setenv("OPCV2_LEAD_PROVIDER", "tianyancha")
	t.Setenv("OPCV2_TYC_API_KEY", "production-tyc-key")
	t.Setenv("OPCV2_CORS_ALLOWED_ORIGINS", "")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want production CORS origin error")
	}
}

func TestLoadRejectsWildcardCORSOriginsInProduction(t *testing.T) {
	t.Setenv("OPCV2_ENV", "production")
	t.Setenv("OPCV2_JWT_SECRET", "production-secret")
	t.Setenv("OPCV2_SMS_PROVIDER", "disabled")
	t.Setenv("OPCV2_AI_PROVIDER", "openai")
	t.Setenv("OPCV2_AI_MODEL", "gpt-production")
	t.Setenv("OPCV2_AI_API_KEY", "production-ai-key")
	t.Setenv("OPCV2_LEAD_PROVIDER", "tianyancha")
	t.Setenv("OPCV2_TYC_API_KEY", "production-tyc-key")
	t.Setenv("OPCV2_CORS_ALLOWED_ORIGINS", "*")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want production wildcard CORS origin error")
	}
}

func TestLoadRejectsProductionWithoutExplicitAIProvider(t *testing.T) {
	t.Setenv("OPCV2_ENV", "production")
	t.Setenv("OPCV2_JWT_SECRET", "production-secret")
	t.Setenv("OPCV2_SMS_PROVIDER", "disabled")
	t.Setenv("OPCV2_AI_PROVIDER", "")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want production AI provider error")
	}
}

func TestLoadRejectsProductionAIProviderWithoutAPIKey(t *testing.T) {
	t.Setenv("OPCV2_ENV", "production")
	t.Setenv("OPCV2_JWT_SECRET", "production-secret")
	t.Setenv("OPCV2_SMS_PROVIDER", "disabled")
	t.Setenv("OPCV2_AI_PROVIDER", "openai")
	t.Setenv("OPCV2_AI_MODEL", "gpt-production")
	t.Setenv("OPCV2_AI_API_KEY", "")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want production AI API key error")
	}
}

func TestLoadRejectsProductionAIProviderWithoutExplicitModel(t *testing.T) {
	t.Setenv("OPCV2_ENV", "production")
	t.Setenv("OPCV2_JWT_SECRET", "production-secret")
	t.Setenv("OPCV2_SMS_PROVIDER", "disabled")
	t.Setenv("OPCV2_AI_PROVIDER", "openai")
	t.Setenv("OPCV2_AI_MODEL", "")
	t.Setenv("OPCV2_AI_API_KEY", "production-ai-key")
	t.Setenv("OPCV2_LEAD_PROVIDER", "tianyancha")
	t.Setenv("OPCV2_TYC_API_KEY", "production-tyc-key")
	t.Setenv("OPCV2_CORS_ALLOWED_ORIGINS", "https://app.example.com")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want production AI model error")
	}
}

func TestLoadRejectsProductionWithoutExplicitLeadProvider(t *testing.T) {
	t.Setenv("OPCV2_ENV", "production")
	t.Setenv("OPCV2_JWT_SECRET", "production-secret")
	t.Setenv("OPCV2_SMS_PROVIDER", "disabled")
	t.Setenv("OPCV2_AI_PROVIDER", "openai")
	t.Setenv("OPCV2_AI_MODEL", "gpt-production")
	t.Setenv("OPCV2_AI_API_KEY", "production-ai-key")
	t.Setenv("OPCV2_LEAD_PROVIDER", "")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want production lead provider error")
	}
}

func TestLoadRejectsProductionTianyanchaWithoutAPIKey(t *testing.T) {
	t.Setenv("OPCV2_ENV", "production")
	t.Setenv("OPCV2_JWT_SECRET", "production-secret")
	t.Setenv("OPCV2_SMS_PROVIDER", "disabled")
	t.Setenv("OPCV2_AI_PROVIDER", "openai")
	t.Setenv("OPCV2_AI_MODEL", "gpt-production")
	t.Setenv("OPCV2_AI_API_KEY", "production-ai-key")
	t.Setenv("OPCV2_LEAD_PROVIDER", "tianyancha")
	t.Setenv("OPCV2_TYC_API_KEY", "")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want production Tianyancha API key error")
	}
}
