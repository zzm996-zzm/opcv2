package main

import (
	"log/slog"
	"os"

	"github.com/zzm/opcv2/internal/platform/config"
	"github.com/zzm/opcv2/internal/platform/taskqueue"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config", "error", err)
		os.Exit(1)
	}

	server := taskqueue.NewServer(cfg.RedisAddr)
	logger.Info("worker starting", "redis_addr", cfg.RedisAddr)
	if err := server.Run(taskqueue.NewMux()); err != nil {
		logger.Error("run worker", "error", err)
		os.Exit(1)
	}
}
