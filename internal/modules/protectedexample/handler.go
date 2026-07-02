package protectedexample

import (
	"net/http"
	"strings"

	"github.com/eecopilot/go-modular-starter/internal/httpserver"
	"github.com/eecopilot/go-modular-starter/internal/modules/auth"
	"github.com/eecopilot/userkit"
)

const Route = "/api/v1/protected/example"

type AuthMiddleware func(http.Handler) http.Handler

type Handler struct {
	auth AuthMiddleware
}

type Response struct {
	CurrentUser userkit.User     `json:"current_user"`
	Resource    BusinessResource `json:"resource"`
}

type BusinessResource struct {
	ID      string `json:"id"`
	OwnerID string `json:"owner_id"`
	Name    string `json:"name"`
}

func New(auth AuthMiddleware) *Handler {
	return &Handler{auth: auth}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	if h.auth == nil {
		return
	}
	mux.Handle("GET "+Route, h.auth(http.HandlerFunc(h.Get)))
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	current, ok := auth.CurrentUser(r)
	if !ok {
		httpserver.WriteError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, Response{
		CurrentUser: current,
		Resource: BusinessResource{
			ID:      "starter-workspace",
			OwnerID: current.ID,
			Name:    "Starter workspace for " + displayName(current),
		},
	})
}

func displayName(user userkit.User) string {
	for _, value := range []string{user.FullName, user.Username, user.Email, user.ID} {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return "current user"
}
