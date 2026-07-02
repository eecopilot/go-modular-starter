package webui

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
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
