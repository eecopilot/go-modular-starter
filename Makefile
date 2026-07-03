.PHONY: dev dev-up dev-down frontend-import frontend-build test build run docker-up docker-app-up docker-down migrate-up docker-build smoke-local smoke-docker

APP_NAME := go-modular-starter
BUILD_DIR := bin
COMPOSE := docker compose -f deployments/docker-compose.yml

.EXPORT_ALL_VARIABLES:

ifneq (,$(wildcard .env))
include .env
endif

USERKIT_DATABASE_URL ?= postgres://starter:starter@localhost:55432/starter?sslmode=disable
USERKIT_JWT_SECRET ?= change-me-to-a-long-random-secret
DATABASE_URL ?= $(USERKIT_DATABASE_URL)

dev:
	go run ./cmd/api

dev-up: export USERKIT_ENABLED := true
dev-up: docker-up migrate-up
	@printf '\napp: http://localhost:8080\n'
	@printf 'health: curl http://localhost:8080/healthz\n'
	@printf 'smoke: make smoke-local\n\n'
	go run ./cmd/api

dev-down: docker-down

frontend-import:
	sh scripts/frontend-import.sh

frontend-build: export FRONTEND_GIT_URL :=
frontend-build: export FRONTEND_GIT_REF :=
frontend-build:
	sh scripts/frontend-import.sh

run:
	go run ./cmd/api

test:
	go test ./...

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/api

docker-build:
	docker build -f deployments/Dockerfile -t $(APP_NAME):local .

docker-up:
	$(COMPOSE) up -d postgres

docker-app-up:
	$(COMPOSE) --profile app up -d --build

docker-down:
	$(COMPOSE) --profile app down

migrate-up:
	$(COMPOSE) --profile app run --rm migrate

smoke-local:
	sh scripts/smoke-http.sh

smoke-docker:
	sh scripts/smoke-http.sh
