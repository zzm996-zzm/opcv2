package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zzm/opcv2/internal/auth"
	"github.com/zzm/opcv2/internal/platform/config"
	"github.com/zzm/opcv2/internal/platform/health"
	"github.com/zzm/opcv2/internal/platform/httpserver"
	"github.com/zzm/opcv2/internal/platform/postgres"
	"github.com/zzm/opcv2/internal/platform/rediscache"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	db, err := postgres.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("open postgres", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	redisClient := rediscache.NewClient(cfg.RedisAddr)
	defer func() { _ = redisClient.Close() }()

	userRepository := auth.NewPostgresUserRepository(db)
	codeStore := auth.NewRedisCodeStore(redisClient.Client)
	sessionStore := auth.NewRedisSessionStore(redisClient.Client)
	tokenManager := auth.NewJWTManager(cfg.JWTSecret, 15*time.Minute)
	if cfg.SMSProvider != "development" {
		logger.Error("unsupported SMS provider", "provider", cfg.SMSProvider)
		os.Exit(1)
	}
	smsProvider := auth.NewDevelopmentSMSProvider(cfg.SMSDevCode)
	authService := auth.NewService(auth.Dependencies{
		Codes:        codeStore,
		SMS:          smsProvider,
		Users:        userRepository,
		Tokens:       tokenManager,
		Sessions:     sessionStore,
		GenerateCode: func() (string, error) { return cfg.SMSDevCode, nil },
	})
	authHTTP := auth.NewHTTPHandler(authService, tokenManager, cfg.Environment == "production")

	checker := health.NewChecker(db, redisClient)
	server := &http.Server{
		Addr: cfg.HTTPAddr,
		Handler: httpserver.NewRouter(httpserver.HealthChecks{
			Ready: func() bool { return checker.Ready(context.Background()) },
		}, authHTTP),
		ReadHeaderTimeout: 5 * time.Second,
	}

	shutdownCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("api listening", "addr", cfg.HTTPAddr, "environment", cfg.Environment)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("serve api", "error", err)
			stop()
		}
	}()

	<-shutdownCtx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("shutdown api", "error", err)
	}
}
