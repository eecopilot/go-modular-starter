package bootstrap

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/eecopilot/go-modular-starter/internal/app"
	"github.com/eecopilot/go-modular-starter/internal/config"
	"github.com/eecopilot/go-modular-starter/internal/httpserver"
	"github.com/eecopilot/go-modular-starter/internal/modules/auth"
	"github.com/eecopilot/go-modular-starter/internal/modules/example"
	"github.com/eecopilot/go-modular-starter/internal/modules/health"
	platformpostgres "github.com/eecopilot/go-modular-starter/internal/platform/postgres"
	"github.com/eecopilot/go-modular-starter/internal/webui"
	"github.com/eecopilot/go-modular-starter/web"
	"github.com/eecopilot/userkit"
	userkitpostgres "github.com/eecopilot/userkit/postgres"
)

func New(cfg config.Config, log *slog.Logger) (*app.App, error) {
	mux := http.NewServeMux()
	var closers []app.Closer
	var readiness []health.NamedChecker

	if cfg.Userkit.Enabled {
		db, err := platformpostgres.Open(context.Background(), platformpostgres.Config{
			DatabaseURL:     cfg.Userkit.DatabaseURL,
			MaxOpenConns:    cfg.Userkit.DBMaxOpenConns,
			MaxIdleConns:    cfg.Userkit.DBMaxIdleConns,
			ConnMaxLifetime: cfg.Userkit.DBConnMaxLifetime,
			ConnMaxIdleTime: cfg.Userkit.DBConnMaxIdleTime,
		})
		if err != nil {
			return nil, err
		}
		closers = append(closers, db)
		readiness = append(readiness, health.NamedChecker{Name: "postgres", Checker: db})

		service, err := userkit.NewService(userkitpostgres.New(db.DB), userkit.Config{
			JWTSecret:   cfg.Userkit.JWTSecret,
			JWTIssuer:   cfg.Userkit.JWTIssuer,
			JWTAudience: cfg.Userkit.JWTAudience,
			TokenTTL:    cfg.Userkit.TokenTTL,
			BCryptCost:  cfg.Userkit.BCryptCost,
		})
		if err != nil {
			_ = db.Close(context.Background())
			return nil, err
		}
		auth.New(service).RegisterRoutes(mux)
		log.Info("userkit module enabled", "prefix", auth.Prefix)
	} else {
		log.Info("userkit module disabled")
	}

	example.New(example.NewMemoryStore()).RegisterRoutes(mux)

	health.Handler{
		Readiness: health.MultiChecker(readiness),
		Version: health.VersionInfo{
			Name:      cfg.App.Name,
			Env:       cfg.App.Env,
			Version:   cfg.App.Version,
			Commit:    cfg.App.Commit,
			BuildTime: cfg.App.BuildTime,
		},
	}.RegisterRoutes(mux)

	webHandler, err := webui.New(web.Files)
	if err != nil {
		return nil, err
	}
	mux.Handle("/", webHandler)

	server := httpserver.New(cfg.HTTP, mux, log)
	return app.New(server, log, closers...), nil
}
