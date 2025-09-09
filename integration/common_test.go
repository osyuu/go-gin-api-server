package integration

import (
	"go-gin-api-server/config"
	"go-gin-api-server/internal/database"
	"go-gin-api-server/internal/model"
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
	db := database.GetDB()
	if err := db.AutoMigrate(&model.User{}, &model.UserCredentials{}); err != nil {
		panic("Failed to migrate test database: " + err.Error())
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
