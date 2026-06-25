.PHONY: help dev build test lint docker-up docker-down

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build:
	cd backend && go build -o ../bin/lil-poker ./cmd/server

test:
	cd backend && go test ./...

lint:
	cd backend && golangci-lint run ./...

frontend-install:
	cd frontend && npm install

frontend-dev:
	cd frontend && npm run dev

frontend-build:
	cd frontend && npm run build

docker-up:
	docker compose up --build

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f

docker-api-logs:
	docker compose logs -f api

dev-backend:
	cd backend && go run ./cmd/server

dev:
	@$(MAKE) -j2 dev-backend frontend-dev
