#!/usr/bin/env bash
set -euo pipefail

echo "==> Installing tools..."
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install golang.org/x/tools/cmd/goimports@latest

echo "==> Downloading Go modules..."
go mod download

echo "==> Starting infrastructure..."
docker compose -f deploy/docker-compose.yml up -d postgres redis

echo "==> Waiting for postgres..."
until docker compose -f deploy/docker-compose.yml exec postgres pg_isready -U user -d service_db 2>/dev/null; do
  echo "  waiting for postgres..."
  sleep 2
done

echo "==> Running migrations..."
bash scripts/migrate.sh up

echo "==> Starting server..."
go run ./cmd/server
