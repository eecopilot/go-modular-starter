.PHONY: dev test build run docker-up docker-app-up docker-down migrate-up docker-build smoke-docker

APP_NAME := go-modular-starter
BUILD_DIR := bin
DATABASE_URL ?= postgres://starter:starter@localhost:55432/starter?sslmode=disable

dev:
	go run ./cmd/api

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
	docker compose -f deployments/docker-compose.yml up -d

docker-app-up:
	docker compose -f deployments/docker-compose.yml --profile app up -d --build

docker-down:
	docker compose -f deployments/docker-compose.yml --profile app down

migrate-up:
	@for file in migrations/*.sql; do \
		echo "applying $$file"; \
		psql "$(DATABASE_URL)" -v ON_ERROR_STOP=1 -f "$$file"; \
	done

smoke-docker:
	sh scripts/smoke-docker.sh
