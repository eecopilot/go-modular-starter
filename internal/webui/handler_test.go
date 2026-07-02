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
