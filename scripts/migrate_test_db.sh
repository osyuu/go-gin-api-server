#!/bin/bash

# Migration 腳本
# 使用方式: ./scripts/migrate.sh [command] [options]
# 例如: ./scripts/migrate.sh down 1

# 載入環境變數
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# 設定預設值
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-password}
DB_NAME=${TEST_DB_NAME:-gin_api_server_test}
DB_SSLMODE=${DB_SSLMODE:-disable}

# 建構資料庫連線字串
DB_URL="postgres://$DB_USER:$DB_PASSWORD@postgres:5432/$DB_NAME?sslmode=$DB_SSLMODE"

# 執行 migrate 指令
docker run --rm \
  --network gin-api-server-test-db_gin-api-test-network \
  -v "$(pwd)/migrations:/migrations" \
  migrate/migrate \
  -path /migrations \
  -database "$DB_URL" \
  "$@"
