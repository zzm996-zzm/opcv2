package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/zzm/opcv2/internal/competitor"
	"github.com/zzm/opcv2/internal/geo"
	"github.com/zzm/opcv2/internal/leads"
	"github.com/zzm/opcv2/internal/platform/config"
	"github.com/zzm/opcv2/internal/platform/postgres"
	"github.com/zzm/opcv2/internal/platform/taskqueue"
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
	geoRepository := geo.NewPostgresRepository(db)
	geoService := geo.NewService(geoRepository)
	competitorRepository := competitor.NewPostgresRepository(db)
	competitorService := competitor.NewService(competitorRepository)
	server := taskqueue.NewServer(cfg.RedisAddr)
	mux := taskqueue.NewMux()
	leads.RegisterWorker(mux, leadsService)
	geo.RegisterWorker(mux, geoService)
	competitor.RegisterWorker(mux, competitorService)
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
