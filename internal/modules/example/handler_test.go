package example

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestExampleCRUD(t *testing.T) {
	store := NewMemoryStore()
	now := time.Date(2026, 7, 2, 15, 0, 0, 0, time.UTC)
	store.now = func() time.Time { return now }

	mux := http.NewServeMux()
	New(store).RegisterRoutes(mux)

	createRR := httptest.NewRecorder()
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/examples", strings.NewReader(`{"name":"First","description":"demo"}`))
	createReq.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", createRR.Code, createRR.Body.String())
	}

	var created Item
	if err := json.NewDecoder(createRR.Body).Decode(&created); err != nil {
		t.Fatalf("decode created item: %v", err)
	}
	if created.ID == "" || created.Name != "First" || created.Description != "demo" {
		t.Fatalf("unexpected created item: %#v", created)
	}

	getRR := httptest.NewRecorder()
	mux.ServeHTTP(getRR, httptest.NewRequest(http.MethodGet, "/api/v1/examples/"+created.ID, nil))
	if getRR.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", getRR.Code, getRR.Body.String())
	}

	next := now.Add(time.Minute)
	store.now = func() time.Time { return next }
	updateRR := httptest.NewRecorder()
	updateReq := httptest.NewRequest(http.MethodPatch, "/api/v1/examples/"+created.ID, strings.NewReader(`{"name":"Updated"}`))
	updateReq.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(updateRR, updateReq)
	if updateRR.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", updateRR.Code, updateRR.Body.String())
	}

	var updated Item
	if err := json.NewDecoder(updateRR.Body).Decode(&updated); err != nil {
		t.Fatalf("decode updated item: %v", err)
	}
	if updated.Name != "Updated" || !updated.UpdatedAt.Equal(next) {
		t.Fatalf("unexpected updated item: %#v", updated)
	}

	listRR := httptest.NewRecorder()
	mux.ServeHTTP(listRR, httptest.NewRequest(http.MethodGet, "/api/v1/examples", nil))
	if listRR.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", listRR.Code, listRR.Body.String())
	}
	if !strings.Contains(listRR.Body.String(), "Updated") {
		t.Fatalf("expected list response to include updated item, got %s", listRR.Body.String())
	}

	deleteRR := httptest.NewRecorder()
	mux.ServeHTTP(deleteRR, httptest.NewRequest(http.MethodDelete, "/api/v1/examples/"+created.ID, nil))
	if deleteRR.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", deleteRR.Code, deleteRR.Body.String())
	}
}

func TestExampleRejectsInvalidInput(t *testing.T) {
	mux := http.NewServeMux()
	New(NewMemoryStore()).RegisterRoutes(mux)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/examples", strings.NewReader(`{"name":" "}`))
	req.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
}
