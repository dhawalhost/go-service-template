#!/usr/bin/env bash
set -euo pipefail

DIRECTION=${1:-up}
DSN=${APP_DATABASE_DSN:-"postgres://user:password@localhost:5432/service_db?sslmode=disable"}
MIGRATIONS_PATH=${APP_DATABASE_MIGRATIONS_PATH:-"./migrations"}

echo "==> Running migration: $DIRECTION"
go run -mod=mod github.com/golang-migrate/migrate/v4/cmd/migrate \
  -path "$MIGRATIONS_PATH" \
  -database "$DSN" \
  "$DIRECTION"
