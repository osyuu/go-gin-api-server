package integration

import (
	"go-gin-api-server/config"
	"go-gin-api-server/internal/database"
	"go-gin-api-server/pkg/logger"
	"os"
	"testing"

	"gorm.io/gorm"
)

// TestMain 設置測試環境
func TestMain(m *testing.M) {
	// 初始化測試專用配置和日誌
	cfg := config.LoadTestConfig()
	logger.Init("test")

	// 初始化測試資料庫
	if err := database.InitDatabase(cfg.Database); err != nil {
		panic("Failed to initialize test database: " + err.Error())
	}

	// 運行資料庫 migration
	// 注意：Integration 測試應該使用 golang-migrate 來管理資料庫結構
	// 請確保在運行 integration 測試前先執行: make migrate-test-up
	db := database.GetDB()
	// 檢查表是否存在，如果不存在則提示用戶運行 migration
	var count int64
	if err := db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'users'").Scan(&count).Error; err != nil {
		panic("Failed to check database tables: " + err.Error())
	}
	if count == 0 {
		panic("Database tables not found. Please run 'make migrate-test-up' before running integration tests.")
	}

	// 運行測試
	code := m.Run()

	// 退出
	os.Exit(code)
}

// setup 為每個測試準備乾淨的資料庫狀態
func setup() *gorm.DB {
	// 清理測試數據
	db := database.GetDB()
	if db != nil {
		db.Exec("DELETE FROM user_credentials")
		db.Exec("DELETE FROM users")
	}
	return db
}

// teardown 測試後的清理處理
func teardown(db *gorm.DB) {
	// 清理測試數據
	if db != nil {
		db.Exec("DELETE FROM user_credentials")
		db.Exec("DELETE FROM users")
	}
}
