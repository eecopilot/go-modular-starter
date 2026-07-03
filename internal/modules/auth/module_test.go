package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRegisterRoutesReturnsJSONNotFoundForUnknownAPIRoute(t *testing.T) {
	mux := http.NewServeMux()
	New(nil).RegisterRoutes(mux)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/unknown", nil))

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected JSON response, got %q", ct)
	}
	if !strings.Contains(rr.Body.String(), "not_found") {
		t.Fatalf("unexpected body: %s", rr.Body.String())
	}
}
