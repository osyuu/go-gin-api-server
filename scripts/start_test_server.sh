#!/bin/bash

# 啟動測試環境服務器腳本

echo "Run test server"
echo "=================================="

# 設置測試環境
export APP_ENV=test

# 設置服務器端口
export PORT=8080

echo "Test environment configuration:"
echo "  Environment: $APP_ENV"
echo "  Database: gin_api_server_test"
echo "  Server port: $PORT"
echo ""

# 檢查測試資料庫是否運行
echo "📡 Check test database..."
if ! docker exec gin-api-postgres psql -U postgres -d gin_api_server_test -c "SELECT 1;" > /dev/null 2>&1; then
    echo "Test database is not running! Please start PostgreSQL:"
    echo "   docker-compose up -d"
    exit 1
fi
echo "Test database is running normally"
echo ""

# 運行資料庫遷移
echo "Run database migration..."
make migrate-test-up
echo ""

# 啟動服務器
echo "Start test environment server..."
echo "Press Ctrl+C to stop the server"
echo ""
go run cmd/server/main.go
