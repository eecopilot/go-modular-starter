package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/eecopilot/go-modular-starter/internal/httpserver"
)

type Closer interface {
	Close(ctx context.Context) error
}

type App struct {
	server  *httpserver.Server
	closers []Closer
	logger  *slog.Logger
}

func New(server *httpserver.Server, logger *slog.Logger, closers ...Closer) *App {
	return &App{server: server, logger: logger, closers: closers}
}

func (a *App) Run(ctx context.Context) error {
	if a.server == nil {
		return fmt.Errorf("app: http server is required")
	}
	err := a.server.Run(ctx)
	if closeErr := a.close(context.Background()); closeErr != nil && err == nil {
		err = closeErr
	}
	return err
}

func (a *App) close(ctx context.Context) error {
	var firstErr error
	for i := len(a.closers) - 1; i >= 0; i-- {
		if err := a.closers[i].Close(ctx); err != nil {
			a.logger.Error("close dependency", "error", err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}
