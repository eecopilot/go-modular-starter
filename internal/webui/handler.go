package webui

import (
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/eecopilot/go-modular-starter/internal/httpserver"
)

type Handler struct {
	files http.Handler
	dist  fs.FS
}

func New(files fs.FS) (*Handler, error) {
	dist, err := fs.Sub(files, "dist")
	if err != nil {
		return nil, err
	}
	return &Handler{
		files: http.FileServerFS(dist),
		dist:  dist,
	}, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// API 路由不匹配时必须返回 JSON 404，绝不能回退到 SPA index。
	if r.URL.Path == "/api" || strings.HasPrefix(r.URL.Path, "/api/") {
		httpserver.WriteError(w, http.StatusNotFound, "not_found", "route not found")
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		httpserver.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "only GET and HEAD are allowed")
		return
	}

	cleanPath := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	if cleanPath == "." || cleanPath == "" {
		h.serveIndex(w)
		return
	}
	if cleanPath == "index.html" {
		h.serveIndex(w)
		return
	}

	info, err := fs.Stat(h.dist, cleanPath)
	if err == nil && !info.IsDir() {
		h.files.ServeHTTP(w, r)
		return
	}

	h.serveIndex(w)
}

func (h *Handler) serveIndex(w http.ResponseWriter) {
	data, err := fs.ReadFile(h.dist, "index.html")
	if err != nil {
		http.Error(w, "frontend index not found", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
