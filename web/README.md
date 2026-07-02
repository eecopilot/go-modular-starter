# Web

This directory contains the starter frontend shell.

Current approach:

- `dist/` contains static assets embedded into the Go binary with `go:embed`.
- No business UI is implemented yet.
- No Node toolchain is required for the initial starter.

Later, if the frontend becomes a full SPA, keep source files under `web/src` and build into `web/dist`.
