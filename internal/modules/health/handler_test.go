package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthRoutes(t *testing.T) {
	mux := http.NewServeMux()
	Handler{
		Readiness: CheckerFunc(func(ctx context.Context) error { return nil }),
		Version:   VersionInfo{Name: "test-api", Version: "dev"},
	}.RegisterRoutes(mux)

	tests := []struct {
		target string
		want   int
		body   string
	}{
		{target: "/healthz", want: http.StatusOK, body: `"ok"`},
		{target: "/readyz", want: http.StatusOK, body: `"ready"`},
		{target: "/version", want: http.StatusOK, body: `"test-api"`},
	}

	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, tt.target, nil))
			if rr.Code != tt.want {
				t.Fatalf("expected %d, got %d", tt.want, rr.Code)
			}
			if !strings.Contains(rr.Body.String(), tt.body) {
				t.Fatalf("expected body to contain %s, got %s", tt.body, rr.Body.String())
			}
		})
	}
}

func TestReadyReturnsUnavailableWhenDependencyFails(t *testing.T) {
	mux := http.NewServeMux()
	Handler{
		Readiness: CheckerFunc(func(ctx context.Context) error { return errors.New("database down") }),
	}.RegisterRoutes(mux)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "database down") {
		t.Fatalf("unexpected body: %s", rr.Body.String())
	}
}
