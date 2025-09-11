package config

import (
	"os"
	"time"
)

type Config struct {
	Env      string
	Port     string
	LogLevel string
	JWT      JWTConfig
	Database DatabaseConfig
}

type JWTConfig struct {
	Secret                 string
	AccessTokenExpiration  time.Duration
	RefreshTokenExpiration time.Duration
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

var AppConfig *Config

func LoadConfig() *Config {
	env := getEnv("APP_ENV", "development")

	// 開發和生產環境配置
	AppConfig = &Config{
		Env:      env,
		Port:     getEnv("PORT", "8080"),
		LogLevel: getEnv("LOG_LEVEL", "debug"),
		JWT: JWTConfig{
			Secret:                 getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
			AccessTokenExpiration:  getDurationEnv("JWT_ACCESS_TOKEN_EXPIRATION", 15*time.Minute),
			RefreshTokenExpiration: getDurationEnv("JWT_REFRESH_TOKEN_EXPIRATION", 7*24*time.Hour),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			DBName:   getEnv("DB_NAME", "gin_api_server"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
	}
	return AppConfig
}

// LoadTestConfig 載入測試專用配置
func LoadTestConfig() *Config {
	return &Config{
		Env:      "test",
		Port:     "8080",
		LogLevel: "error", // 測試時減少日誌輸出
		JWT: JWTConfig{
			Secret:                 "test-secret-key",
			AccessTokenExpiration:  15 * time.Minute,
			RefreshTokenExpiration: 7 * 24 * time.Hour,
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("TEST_DB_PORT", "5433"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			DBName:   getEnv("TEST_DB_NAME", "gin_api_server_test"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return fallback
}
