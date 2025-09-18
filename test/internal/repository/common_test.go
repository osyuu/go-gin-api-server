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
	logger.Init(config.Test)

	logger.Log.Info("TestMain started - Initializing repository test environment")

	// Initialize test database
	if err := database.InitDatabase(cfg.Database); err != nil {
		panic("Failed to initialize test database: " + err.Error())
	}

	// Run database migration
	// Note: Integration tests should use golang-migrate to manage database schema
	// Please ensure to run: make migrate-test-up before running integration tests
	db := database.GetDB()
	// Check if tables exist, if not, prompt user to run migration
	var count int64
	if err := db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'users'").Scan(&count).Error; err != nil {
		panic("Failed to check database tables: " + err.Error())
	}
	if count == 0 {
		panic("Database tables not found. Please ensure test database is set up with migrations.")
	}

	// Run tests
	logger.Log.Info("Running repository tests...")
	code := m.Run()

	logger.Log.Info("TestMain completed - Running tests")

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
