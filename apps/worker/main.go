package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/zzm/opcv2/internal/account"
	"github.com/zzm/opcv2/internal/ai"
	"github.com/zzm/opcv2/internal/competitor"
	"github.com/zzm/opcv2/internal/geo"
	"github.com/zzm/opcv2/internal/leads"
	"github.com/zzm/opcv2/internal/membership"
	"github.com/zzm/opcv2/internal/platform/aiprovider"
	"github.com/zzm/opcv2/internal/platform/config"
	"github.com/zzm/opcv2/internal/platform/postgres"
	"github.com/zzm/opcv2/internal/platform/taskqueue"
	"github.com/zzm/opcv2/internal/sandbox"
	"github.com/zzm/opcv2/internal/tasks"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config", "error", err)
		os.Exit(1)
	}

	db, err := postgres.Open(context.Background(), cfg.DatabaseURL)
	if err != nil {
		logger.Error("open postgres", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	membershipRepository := membership.NewPostgresRepository(db)
	membershipService := membership.NewService(membershipRepository)
	aiProvider, err := aiprovider.New(cfg)
	if err != nil {
		logger.Error("configure ai provider", "provider", cfg.AIProvider, "error", err)
		os.Exit(1)
	}
	aiService := ai.NewService(ai.NewPostgresRepository(db), aiProvider, ai.Config{Provider: cfg.AIProvider, Model: cfg.AIModel, Logger: logger})
	accountService := account.NewService(account.NewPostgresRepository(db))
	leadsRepository := leads.NewPostgresRepository(db)
	leadProvider, err := newLeadProvider(cfg)
	if err != nil {
		logger.Error("configure lead provider", "provider", cfg.LeadProvider, "error", err)
		os.Exit(1)
	}
	leadsService := leads.NewService(
		leadsRepository,
		membershipService,
		taskqueue.NewClient(cfg.RedisAddr),
		leadProvider,
	)
	geoRepository := geo.NewPostgresRepository(db)
	geoService := geo.NewService(geoRepository)
	competitorRepository := competitor.NewPostgresRepository(db)
	competitorScanner, err := newCompetitorScanner(cfg)
	if err != nil {
		logger.Error("configure competitor scanner", "provider", cfg.CompetitorScannerProvider, "error", err)
		os.Exit(1)
	}
	competitorOptions := []competitor.Option{}
	if competitorScanner != nil {
		competitorOptions = append(competitorOptions, competitor.WithScanner(competitorScanner))
	}
	competitorService := competitor.NewService(competitorRepository, competitorOptions...)
	sandboxService := sandbox.NewService(sandbox.NewPostgresRepository(db), aiService, sandbox.WithQuotaConsumer(membershipService), sandbox.WithProfileContextProvider(accountService))
	tasksRepository := tasks.NewPostgresRepository(db)
	tasksService := tasks.NewService(tasksRepository)
	go tasks.RunReminderWorker(context.Background(), tasksService, time.Minute, func(err error) {
		logger.Error("dispatch task reminders", "error", err)
	})
	server := taskqueue.NewServer(cfg.RedisAddr)
	mux := taskqueue.NewMux()
	leads.RegisterWorker(mux, leadsService)
	geo.RegisterWorker(mux, geoService)
	competitor.RegisterWorker(mux, competitorService)
	sandbox.RegisterWorker(mux, sandboxService)
	logger.Info("worker starting", "redis_addr", cfg.RedisAddr)
	if err := server.Run(mux); err != nil {
		logger.Error("run worker", "error", err)
		os.Exit(1)
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

func newCompetitorScanner(cfg config.Config) (competitor.Scanner, error) {
	switch cfg.CompetitorScannerProvider {
	case "":
		return nil, nil
	case "development":
		return competitor.NewDevelopmentScanner(), nil
	default:
		return nil, errors.New("unsupported competitor scanner provider")
	}
}
