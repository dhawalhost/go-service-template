.PHONY: run build test lint fmt vet tidy migrate-up migrate-down docker-build docker-run help

BINARY=./bin/server

run: ## Run the service locally
	go run ./cmd/server

build: ## Build the binary
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BINARY) ./cmd/server

test: ## Run tests with race detector
	go test ./... -race -cover

lint: ## Run golangci-lint
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest run --tests=false --timeout=15m ./...

fmt: ## Format code
	gofmt -w .

vet: ## Run go vet
	go vet ./...

tidy: ## Tidy go modules
	go mod tidy

migrate-up: ## Run database migrations up
	@bash scripts/migrate.sh up

migrate-down: ## Run database migrations down
	@bash scripts/migrate.sh down

docker-build: ## Build Docker image
	docker build -f deploy/Dockerfile -t go-service-template:latest .

docker-run: ## Start full local stack via docker compose
	docker compose -f deploy/docker-compose.yml up --build

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
