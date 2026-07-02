package protectedexample

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eecopilot/go-modular-starter/internal/modules/auth"
	"github.com/eecopilot/userkit"
)

func TestProtectedExampleReturnsCurrentUser(t *testing.T) {
	current := userkit.User{
		ID:       "user-123",
		Email:    "alice@example.com",
		Username: "alice",
		FullName: "Alice Example",
		Status:   userkit.UserStatusEnabled,
	}

	mux := http.NewServeMux()
	New(withUser(current)).RegisterRoutes(mux)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, Route, nil)
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var response Response
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.CurrentUser.ID != current.ID {
		t.Fatalf("expected current user %q, got %q", current.ID, response.CurrentUser.ID)
	}
	if response.Resource.OwnerID != current.ID {
		t.Fatalf("expected resource owner %q, got %q", current.ID, response.Resource.OwnerID)
	}
	if response.Resource.Name != "Starter workspace for Alice Example" {
		t.Fatalf("unexpected resource name %q", response.Resource.Name)
	}
}

func TestProtectedExampleRequiresCurrentUser(t *testing.T) {
	mux := http.NewServeMux()
	New(passThrough).RegisterRoutes(mux)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, Route, nil)
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rr.Code, rr.Body.String())
	}
}

func withUser(user userkit.User) AuthMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(auth.ContextWithUser(r.Context(), user)))
		})
	}
}

func passThrough(next http.Handler) http.Handler {
	return next
}
