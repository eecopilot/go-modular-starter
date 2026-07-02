package example

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/eecopilot/go-modular-starter/internal/httpserver"
)

type Item struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateItemInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateItemInput struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type Store interface {
	List() []Item
	Create(input CreateItemInput) (Item, error)
	Get(id string) (Item, bool)
	Update(id string, input UpdateItemInput) (Item, error)
	Delete(id string) bool
}

type MemoryStore struct {
	mu    sync.RWMutex
	items map[string]Item
	now   func() time.Time
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		items: make(map[string]Item),
		now:   time.Now,
	}
}

func (s *MemoryStore) List() []Item {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]Item, 0, len(s.items))
	for _, item := range s.items {
		items = append(items, item)
	}
	return items
}

func (s *MemoryStore) Create(input CreateItemInput) (Item, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return Item{}, errors.New("name is required")
	}

	now := s.now()
	item := Item{
		ID:          newID(),
		Name:        name,
		Description: strings.TrimSpace(input.Description),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[item.ID] = item
	return item, nil
}

func (s *MemoryStore) Get(id string) (Item, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.items[id]
	return item, ok
}

func (s *MemoryStore) Update(id string, input UpdateItemInput) (Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.items[id]
	if !ok {
		return Item{}, errNotFound
	}
	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return Item{}, errors.New("name is required")
		}
		item.Name = name
	}
	if input.Description != nil {
		item.Description = strings.TrimSpace(*input.Description)
	}
	item.UpdatedAt = s.now()
	s.items[id] = item
	return item, nil
}

func (s *MemoryStore) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.items[id]; !ok {
		return false
	}
	delete(s.items, id)
	return true
}

type Handler struct {
	store Store
}

func New(store Store) *Handler {
	if store == nil {
		store = NewMemoryStore()
	}
	return &Handler{store: store}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/examples", h.List)
	mux.HandleFunc("POST /api/v1/examples", h.Create)
	mux.HandleFunc("GET /api/v1/examples/{id}", h.Get)
	mux.HandleFunc("PATCH /api/v1/examples/{id}", h.Update)
	mux.HandleFunc("DELETE /api/v1/examples/{id}", h.Delete)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{"items": h.store.List()})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input CreateItemInput
	if err := decodeJSON(r, &input); err != nil {
		httpserver.WriteError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}
	item, err := h.store.Create(input)
	if err != nil {
		httpserver.WriteError(w, http.StatusBadRequest, "invalid_input", err.Error())
		return
	}
	httpserver.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	item, ok := h.store.Get(r.PathValue("id"))
	if !ok {
		httpserver.WriteError(w, http.StatusNotFound, "not_found", "example item not found")
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	var input UpdateItemInput
	if err := decodeJSON(r, &input); err != nil {
		httpserver.WriteError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}
	item, err := h.store.Update(r.PathValue("id"), input)
	if errors.Is(err, errNotFound) {
		httpserver.WriteError(w, http.StatusNotFound, "not_found", "example item not found")
		return
	}
	if err != nil {
		httpserver.WriteError(w, http.StatusBadRequest, "invalid_input", err.Error())
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	if !h.store.Delete(r.PathValue("id")) {
		httpserver.WriteError(w, http.StatusNotFound, "not_found", "example item not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

var errNotFound = errors.New("example item not found")

func decodeJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

func newID() string {
	var data [12]byte
	if _, err := rand.Read(data[:]); err != nil {
		return hex.EncodeToString([]byte(time.Now().Format("20060102150405.000000000")))
	}
	return hex.EncodeToString(data[:])
}
