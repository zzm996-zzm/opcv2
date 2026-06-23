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
	t.Setenv("OPCV2_AUTO_MIGRATE", "")
	t.Setenv("OPCV2_MIGRATIONS_PATH", "")

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
}

func TestLoadUsesEnvironmentOverrides(t *testing.T) {
	t.Setenv("OPCV2_ENV", "test")
	t.Setenv("OPCV2_HTTP_ADDR", ":9090")
	t.Setenv("OPCV2_DATABASE_URL", "postgres://example")
	t.Setenv("OPCV2_REDIS_ADDR", "redis:6380")
	t.Setenv("OPCV2_JWT_SECRET", "test-jwt-secret")
	t.Setenv("OPCV2_SMS_PROVIDER", "development")
	t.Setenv("OPCV2_SMS_DEV_CODE", "123456")
	t.Setenv("OPCV2_AUTO_MIGRATE", "false")
	t.Setenv("OPCV2_MIGRATIONS_PATH", "/app/migrations")

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
