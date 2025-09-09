package repository

import (
	"go-gin-api-server/config"
	"go-gin-api-server/internal/database"
	"go-gin-api-server/internal/model"
	"go-gin-api-server/pkg/logger"
	"os"
	"testing"

	"gorm.io/gorm"
)

func TestMain(m *testing.M) {
	// Initialize test configuration and logger
	cfg := config.LoadTestConfig()
	logger.Init("test")

	// Initialize test database
	if err := database.InitDatabase(cfg.Database); err != nil {
		panic("Failed to initialize test database: " + err.Error())
	}

	// Run database migration
	db := database.GetDB()
	if err := db.AutoMigrate(&model.User{}, &model.UserCredentials{}); err != nil {
		panic("Failed to migrate test database: " + err.Error())
	}

	// Run tests
	code := m.Run()

	// Exit
	os.Exit(code)
}

// setup for each test
func setup() *gorm.DB {
	// Start a new transaction, will be rolled back after the test
	db := database.GetDB()
	return db.Begin()
}

// teardown after the test
func teardown(tx *gorm.DB) {
	// Rollback the transaction, will automatically clean up all changes
	tx.Rollback()
}
