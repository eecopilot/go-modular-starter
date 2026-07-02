package logger

import (
	"log/slog"
	"os"

	"github.com/eecopilot/go-modular-starter/internal/config"
)

func New(cfg config.LogConfig) *slog.Logger {
	opts := &slog.HandlerOptions{Level: level(cfg.Level)}
	if cfg.Format == "json" {
		return slog.New(slog.NewJSONHandler(os.Stdout, opts))
	}
	return slog.New(slog.NewTextHandler(os.Stdout, opts))
}

func level(value string) slog.Level {
	switch value {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
