package web

import "embed"

// Files contains the built frontend assets served by the Go binary.
//
//go:embed dist/*
var Files embed.FS
