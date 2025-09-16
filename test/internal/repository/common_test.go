package repository

import (
	"go-gin-api-server/config"
	"go-gin-api-server/internal/database"
	"go-gin-api-server/pkg/logger"
	"os"
	"testing"

	"gorm.io/gorm"
)

const (
	NonExistentUserID = "550e8400-e29b-41d4-a716-446655440000"
	NonExistentPostID = 999999999999
	InvalidCursorID   = "invalid-id"
)

func TestMain(m *testing.M) {
	// Initialize test configuration and logger
	cfg := config.LoadTestConfig()
	logger.Init("test")

	// Initialize test database
	if err := database.InitDatabase(cfg.Database); err != nil {
		panic("Failed to initialize test database: " + err.Error())
	}

	// Run database migration using golang-migrate
	// Note: Test database should be set up with proper migrations before running tests
	// For now, we'll skip AutoMigrate since we're using golang-migrate
	// db := database.GetDB()
	// if err := db.AutoMigrate(&model.User{}, &model.UserCredentials{}); err != nil {
	// 	panic("Failed to migrate test database: " + err.Error())
	// }

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
