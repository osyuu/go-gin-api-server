#!/bin/bash

# Load Balancer 測試腳本
# 簡單測試 Nginx Load Balancer 功能

set -e

# 顏色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 配置
NGINX_URL="http://localhost"

echo -e "${YELLOW}=== Load Balancer 測試 ===${NC}"
echo

# 檢查 Docker 容器狀態
echo "1. 檢查 Docker 容器狀態..."
if docker-compose ps | grep -q "Up"; then
    echo -e "  Docker 容器: ${GREEN}✓${NC}"
    docker-compose ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}"
else
    echo -e "  Docker 容器: ${RED}✗${NC}"
    echo "請啟動: docker-compose up -d"
    exit 1
fi
echo

# 檢查 Nginx
echo "2. 檢查 Nginx..."
if curl -s -f "$NGINX_URL/nginx-health" > /dev/null 2>&1; then
    echo -e "  Nginx: ${GREEN}✓${NC}"
else
    echo -e "  Nginx: ${RED}✗${NC}"
    echo "請啟動: docker-compose up -d"
    exit 1
fi
echo

# 負載測試
echo "3. 負載測試..."

# 檢查是否安裝了 hey
if ! command -v hey &> /dev/null; then
    echo -e "${RED}錯誤: 未安裝 hey${NC}"
    echo "請安裝: brew install hey"
    exit 1
fi

echo "使用 hey 進行負載測試..."
hey -n 1000 -c 10 "$NGINX_URL/health"
echo

echo
echo -e "${GREEN}測試完成！${NC}"