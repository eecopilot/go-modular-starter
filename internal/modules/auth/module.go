package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/eecopilot/go-modular-starter/internal/httpserver"
	"github.com/eecopilot/userkit"
	"github.com/eecopilot/userkit/httpadapter"
)

const Prefix = "/api/v1"

type contextKey struct{}

type Module struct {
	service *userkit.Service
	handler *httpadapter.Handler
}

func New(service *userkit.Service) *Module {
	return &Module{
		service: service,
		handler: httpadapter.New(service),
	}
}

func (m *Module) RegisterRoutes(mux *http.ServeMux) {
	sub := http.NewServeMux()
	m.handler.RegisterRoutes(sub)

	mux.Handle(Prefix+"/", http.StripPrefix(Prefix, sub))
}

func (m *Module) RequireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := m.service.ValidateToken(r.Context(), bearerToken(r.Header.Get("Authorization")))
		if err != nil {
			writeAuthError(w, err)
			return
		}

		next.ServeHTTP(w, r.WithContext(ContextWithUser(r.Context(), user)))
	})
}

func CurrentUser(r *http.Request) (userkit.User, bool) {
	user, ok := r.Context().Value(contextKey{}).(userkit.User)
	return user, ok
}

func ContextWithUser(ctx context.Context, user userkit.User) context.Context {
	return context.WithValue(ctx, contextKey{}, user)
}

func bearerToken(header string) string {
	prefix := "Bearer "
	if len(header) < len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return ""
	}
	return strings.TrimSpace(header[len(prefix):])
}

func writeAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, userkit.ErrUnauthorized),
		errors.Is(err, userkit.ErrInvalidCredentials),
		errors.Is(err, userkit.ErrDisabledUser):
		httpserver.WriteError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
	case errors.Is(err, userkit.ErrForbidden):
		httpserver.WriteError(w, http.StatusForbidden, "forbidden", "permission denied")
	default:
		httpserver.WriteError(w, http.StatusInternalServerError, "internal_error", "authentication failed")
	}
}
