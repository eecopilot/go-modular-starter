package health

import (
	"context"
	"fmt"
	"net/http"

	"github.com/eecopilot/go-modular-starter/internal/httpserver"
)

type Checker interface {
	Check(ctx context.Context) error
}

type CheckerFunc func(ctx context.Context) error

func (f CheckerFunc) Check(ctx context.Context) error {
	return f(ctx)
}

type NamedChecker struct {
	Name    string
	Checker Checker
}

type MultiChecker []NamedChecker

func (m MultiChecker) Check(ctx context.Context) error {
	for _, item := range m {
		if item.Checker == nil {
			continue
		}
		if err := item.Checker.Check(ctx); err != nil {
			if item.Name == "" {
				return err
			}
			return fmt.Errorf("%s: %w", item.Name, err)
		}
	}
	return nil
}

type VersionInfo struct {
	Name      string `json:"name"`
	Env       string `json:"env"`
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"build_time"`
}

type Handler struct {
	Readiness Checker
	Version   VersionInfo
}

func (h Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", h.Health)
	mux.HandleFunc("GET /readyz", h.Ready)
	mux.HandleFunc("GET /version", h.VersionInfo)
}

func (h Handler) Health(w http.ResponseWriter, r *http.Request) {
	httpserver.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h Handler) Ready(w http.ResponseWriter, r *http.Request) {
	if h.Readiness != nil {
		if err := h.Readiness.Check(r.Context()); err != nil {
			httpserver.WriteError(w, http.StatusServiceUnavailable, "not_ready", err.Error())
			return
		}
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (h Handler) VersionInfo(w http.ResponseWriter, r *http.Request) {
	httpserver.WriteJSON(w, http.StatusOK, h.Version)
}
