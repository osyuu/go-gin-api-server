package database

import (
	"fmt"
	"time"

	"go-gin-api-server/config"
	"go-gin-api-server/pkg/logger"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDatabase 初始化資料庫連接
func InitDatabase(cfg config.DatabaseConfig) error {

	// 構建 DSN (Data Source Name)
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s timezone=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode, "UTC")

	// 配置 GORM
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.NewGormLogger(),
		NowFunc: func() time.Time {
			return time.Now().UTC().Truncate(time.Microsecond)
		},
	})

	if err != nil {
		return err
	}

	// 測試連接
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	if err := sqlDB.Ping(); err != nil {
		return err
	}

	logger.Log.Info("Database connected successfully")
	return nil
}

// CloseDatabase 關閉資料庫連接
func CloseDatabase() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	if err := sqlDB.Close(); err != nil {
		return err
	}

	logger.Log.Info("Database connection closed")
	return nil
}

// GetDB 獲取資料庫實例
func GetDB() *gorm.DB {
	return DB
}
