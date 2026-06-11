package config

import "testing"

func TestLoadUsesDevelopmentDefaults(t *testing.T) {
	t.Setenv("OPCV2_ENV", "")
	t.Setenv("OPCV2_HTTP_ADDR", "")
	t.Setenv("OPCV2_DATABASE_URL", "")
	t.Setenv("OPCV2_REDIS_ADDR", "")

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
}

func TestLoadUsesEnvironmentOverrides(t *testing.T) {
	t.Setenv("OPCV2_ENV", "test")
	t.Setenv("OPCV2_HTTP_ADDR", ":9090")
	t.Setenv("OPCV2_DATABASE_URL", "postgres://example")
	t.Setenv("OPCV2_REDIS_ADDR", "redis:6380")

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
}
