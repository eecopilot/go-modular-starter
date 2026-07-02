package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/eecopilot/go-modular-starter/internal/bootstrap"
	"github.com/eecopilot/go-modular-starter/internal/config"
	"github.com/eecopilot/go-modular-starter/internal/platform/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	log := logger.New(cfg.Log)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app, err := bootstrap.New(cfg, log)
	if err != nil {
		log.Error("bootstrap app", "error", err)
		os.Exit(1)
	}

	if err := app.Run(ctx); err != nil {
		log.Error("app stopped with error", "error", err)
		os.Exit(1)
	}
}
