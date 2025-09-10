# 資料庫遷移指南

本專案使用 `golang-migrate` 套件來管理資料庫遷移，提供版本控制和可回滾的資料庫結構變更。

## 架構說明

### 分離式設計

- **`cmd/server/main.go`**: 主應用程式伺服器
- **`cmd/migrate/main.go`**: 獨立的資料庫遷移工具
- **`migrations/`**: 存放 SQL 遷移檔案的目錄

### 優點

1. **關注點分離**: 遷移工具與主應用程式分離
2. **獨立執行**: 可以在不啟動 API 伺服器的情況下執行遷移
3. **版本控制**: 每個遷移都有版本號，可以追蹤變更歷史
4. **可回滾**: 支援向上和向下遷移
5. **團隊協作**: 遷移檔案可以納入版本控制，團隊成員可以同步資料庫結構

## 使用方法

### 基本命令

```bash
# 執行所有待處理的遷移
make migrate-up

# 回滾所有遷移
make migrate-down

# 回滾一個遷移
make migrate-down-1

# 執行一個遷移
make migrate-up-1

# 查看當前遷移版本
make migrate-version

# 強制設定遷移版本（用於修復損壞的遷移狀態）
make migrate-force VERSION=1
```

### 直接使用 Go 命令

```bash
# 執行所有遷移
go run cmd/migrate/main.go -action=up

# 回滾一個遷移
go run cmd/migrate/main.go -action=down -steps=1

# 查看版本
go run cmd/migrate/main.go -action=version
```

## 遷移檔案結構

遷移檔案遵循以下命名規則：

- `{version}_{description}.up.sql` - 向上遷移
- `{version}_{description}.down.sql` - 向下遷移

例如：

- `001_create_users_table.up.sql`
- `001_create_users_table.down.sql`

## 現有遷移檔案

1. **001_create_users_table**: 創建 users 表
2. **002_create_user_credentials_table**: 創建 user_credentials 表
3. **003_create_posts_table**: 創建 posts 表

## 創建新遷移

### 手動創建

1. 在 `migrations/` 目錄下創建新的遷移檔案
2. 使用下一個版本號（例如：004\_）
3. 創建 `.up.sql` 和 `.down.sql` 檔案

### 使用腳本創建（可選）

```bash
# 創建新的遷移檔案
./scripts/create_migration.sh add_user_profile_table
```

## 最佳實踐

1. **總是創建 down 遷移**: 確保每個 up 遷移都有對應的 down 遷移
2. **測試遷移**: 在開發環境中測試 up 和 down 遷移
3. **備份資料**: 在生產環境執行遷移前先備份資料
4. **小步驟**: 將大的變更拆分成多個小的遷移
5. **不可變**: 一旦遷移被部署到生產環境，不要修改它

## 環境變數

遷移工具使用與主應用程式相同的資料庫配置：

```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=gin_api_server
DB_SSLMODE=disable
```

## 故障排除

### 遷移狀態損壞

如果遷移狀態損壞，可以使用 force 命令修復：

```bash
make migrate-force VERSION=1
```

### 檢查遷移狀態

```bash
make migrate-version
```

### 查看遷移歷史

```bash
# 連接到資料庫查看 schema_migrations 表
psql -h localhost -U postgres -d gin_api_server -c "SELECT * FROM schema_migrations;"
```

## 與 GORM 的關係

- 本專案已將 GORM 的 `AutoMigrate` 功能標記為棄用
- 建議使用 SQL 遷移檔案來管理資料庫結構變更
- GORM 仍然用於 ORM 操作，但不用於結構遷移
