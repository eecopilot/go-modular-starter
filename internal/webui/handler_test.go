package webui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/eecopilot/go-modular-starter/web"
)

func TestHandlerServesEmbeddedIndex(t *testing.T) {
	handler, err := New(web.Files)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Go Modular Starter") {
		t.Fatalf("unexpected body: %s", rr.Body.String())
	}
}

func TestHandlerServesAssets(t *testing.T) {
	handler, err := New(web.Files)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/assets/app.js", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "dataset.ready") {
		t.Fatalf("unexpected body: %s", rr.Body.String())
	}
}

func TestHandlerFallsBackToIndexForSPARoutes(t *testing.T) {
	handler, err := New(web.Files)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/dashboard/settings", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Web server ready") {
		t.Fatalf("unexpected body: %s", rr.Body.String())
	}
}

func TestHandlerReturnsJSONNotFoundForAPIRoutes(t *testing.T) {
	handler, err := New(web.Files)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	for _, target := range []string{"/api", "/api/", "/api/v1/unknown"} {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, target, nil))

		if rr.Code != http.StatusNotFound {
			t.Fatalf("%s: expected 404, got %d", target, rr.Code)
		}
		if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
			t.Fatalf("%s: expected JSON response, got %q", target, ct)
		}
		if !strings.Contains(rr.Body.String(), "not_found") {
			t.Fatalf("%s: unexpected body: %s", target, rr.Body.String())
		}
	}
}

func TestHandlerRejectsNonGetMethods(t *testing.T) {
	handler, err := New(web.Files)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/dashboard/settings", nil))

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
	if allow := rr.Header().Get("Allow"); allow != "GET, HEAD" {
		t.Fatalf("expected Allow header, got %q", allow)
	}
}
