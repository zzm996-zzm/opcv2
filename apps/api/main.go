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

	"github.com/zzm/opcv2/internal/ai"
	"github.com/zzm/opcv2/internal/analysis"
	"github.com/zzm/opcv2/internal/auth"
	"github.com/zzm/opcv2/internal/content"
	"github.com/zzm/opcv2/internal/crm"
	"github.com/zzm/opcv2/internal/leads"
	"github.com/zzm/opcv2/internal/membership"
	"github.com/zzm/opcv2/internal/platform/config"
	"github.com/zzm/opcv2/internal/platform/health"
	"github.com/zzm/opcv2/internal/platform/httpserver"
	"github.com/zzm/opcv2/internal/platform/migrations"
	"github.com/zzm/opcv2/internal/platform/postgres"
	"github.com/zzm/opcv2/internal/platform/rediscache"
	"github.com/zzm/opcv2/internal/platform/taskqueue"
	"github.com/zzm/opcv2/internal/projects"
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

	if cfg.AutoMigrate {
		result, err := migrations.Run(ctx, db, cfg.MigrationsPath)
		if err != nil {
			logger.Error("run migrations", "path", cfg.MigrationsPath, "error", err)
			os.Exit(1)
		}
		logger.Info("migrations complete", "path", cfg.MigrationsPath, "current_version", result.CurrentVersion, "applied_count", len(result.Applied))
	}

	redisClient := rediscache.NewClient(cfg.RedisAddr)
	defer func() { _ = redisClient.Close() }()

	userRepository := auth.NewPostgresUserRepository(db)
	codeStore := auth.NewRedisCodeStore(redisClient.Client)
	sessionStore := auth.NewRedisSessionStore(redisClient.Client)
	tokenManager := auth.NewJWTManager(cfg.JWTSecret, 15*time.Minute)
	var smsProvider auth.SMSProvider
	switch cfg.SMSProvider {
	case "development":
		smsProvider = auth.NewDevelopmentSMSProvider(cfg.SMSDevCode)
	case "disabled":
		smsProvider = auth.DisabledSMSProvider{}
	default:
		logger.Error("unsupported SMS provider", "provider", cfg.SMSProvider)
		os.Exit(1)
	}
	authService := auth.NewService(auth.Dependencies{
		Codes:        codeStore,
		SMS:          smsProvider,
		Users:        userRepository,
		Tokens:       tokenManager,
		Sessions:     sessionStore,
		GenerateCode: func() (string, error) { return cfg.SMSDevCode, nil },
	})
	authHTTP := auth.NewHTTPHandler(authService, tokenManager, cfg.Environment == "production")
	membershipRepository := membership.NewPostgresRepository(db)
	membershipService := membership.NewService(membershipRepository)
	membershipHTTP := membership.NewHTTPHandler(membershipService)
	aiRepository := ai.NewPostgresRepository(db)
	aiProvider, err := newAIProvider(cfg)
	if err != nil {
		logger.Error("configure ai provider", "provider", cfg.AIProvider, "error", err)
		os.Exit(1)
	}
	aiService := ai.NewService(aiRepository, aiProvider, ai.Config{
		Provider: cfg.AIProvider,
		Model:    cfg.AIModel,
		Logger:   logger,
	})
	analysisRepository := analysis.NewPostgresRepository(db)
	analysisService := analysis.NewService(analysisRepository, aiService)
	analysisHTTP := analysis.NewHTTPHandler(analysisService)
	projectsRepository := projects.NewPostgresRepository(db)
	projectsService := projects.NewService(projectsRepository, aiService)
	projectsHTTP := projects.NewHTTPHandler(projectsService)
	leadsRepository := leads.NewPostgresRepository(db)
	leadProvider, err := newLeadProvider(cfg)
	if err != nil {
		logger.Error("configure lead provider", "provider", cfg.LeadProvider, "error", err)
		os.Exit(1)
	}
	leadsService := leads.NewService(
		leadsRepository,
		leads.NewDevelopmentCreditLedger(),
		taskqueue.NewClient(cfg.RedisAddr),
		leadProvider,
	)
	leadsHTTP := leads.NewHTTPHandler(leadsService)
	crmRepository := crm.NewPostgresRepository(db)
	crmService := crm.NewService(crmRepository, aiService)
	crmHTTP := crm.NewHTTPHandler(crmService)
	contentRepository := content.NewPostgresRepository(db)
	contentService := content.NewService(contentRepository)
	contentHTTP := content.NewHTTPHandler(contentService)

	checker := health.NewChecker(db, redisClient)
	server := &http.Server{
		Addr: cfg.HTTPAddr,
		Handler: httpserver.NewRouter(httpserver.HealthChecks{
			Ready:                  func() bool { return checker.Ready(context.Background()) },
			Logger:                 logger,
			AllowedOrigins:         cfg.CORSAllowedOrigins,
			ExpensiveEndpointLimit: cfg.ExpensiveEndpointLimit,
		}, authHTTP, membershipHTTP, analysisHTTP, projectsHTTP, leadsHTTP, crmHTTP, contentHTTP),
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

func newLeadProvider(cfg config.Config) (leads.LeadProvider, error) {
	switch cfg.LeadProvider {
	case "development":
		return leads.NewDevelopmentProvider(), nil
	case "tianyancha":
		return leads.NewTianyanchaProvider(leads.TianyanchaConfig{
			BaseURL: cfg.TianyanchaBaseURL,
			APIKey:  cfg.TianyanchaAPIKey,
			Timeout: time.Duration(cfg.TianyanchaTimeoutSeconds) * time.Second,
		}), nil
	default:
		return nil, errors.New("unsupported lead provider")
	}
}

func newAIProvider(cfg config.Config) (ai.Provider, error) {
	switch cfg.AIProvider {
	case "development":
		return ai.NewDevelopmentProvider(), nil
	default:
		return nil, errors.New("unsupported AI provider")
	}
}
