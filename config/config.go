package config

import (
	"log"
	"net/url"
	"os"
	"time"
)

const (
	Development = "development"
	Production  = "production"
	Test        = "test"
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
	URL      string
}

var AppConfig *Config

func LoadConfig() *Config {
	env := getEnv("APP_ENV", Development)
	databaseURL := getEnv("DATABASE_URL", "")
	dbConfig := parseDatabaseURL(databaseURL)

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
		Database: dbConfig,
	}

	// 生產環境安全檢查
	if AppConfig.Env == Production {
		validateProductionConfig(AppConfig)
	}

	return AppConfig
}

// LoadTestConfig 載入測試專用配置
func LoadTestConfig() *Config {
	return &Config{
		Env:      Test,
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

// parseDatabaseURL 解析 DATABASE_URL
func parseDatabaseURL(databaseURL string) DatabaseConfig {
	u, err := url.Parse(databaseURL)
	if err != nil {
		return DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			DBName:   getEnv("DB_NAME", "gin_api_server"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		}
	}

	password, _ := u.User.Password()
	sslMode := "require"
	if u.Query().Get("sslmode") != "" {
		sslMode = u.Query().Get("sslmode")
	}

	port := u.Port()
	if port == "" {
		port = "5432"
	}

	dbName := u.Path[1:] // remove the leading "/"
	if dbName == "" {
		dbName = "gin_api_server"
	}

	return DatabaseConfig{
		Host:     u.Hostname(),
		Port:     port,
		User:     u.User.Username(),
		Password: password,
		DBName:   dbName,
		SSLMode:  sslMode,
		URL:      databaseURL,
	}
}

// validateProductionConfig 驗證生產環境配置
func validateProductionConfig(cfg *Config) {
	// 檢查 JWT Secret 是否為預設值
	if cfg.JWT.Secret == "your-secret-key-change-in-production" {
		log.Fatal("JWT_SECRET must be changed in production")
	}

	// 檢查 JWT Secret 長度
	if len(cfg.JWT.Secret) < 32 {
		log.Fatal("JWT_SECRET must be at least 32 characters in production")
	}

	// 檢查資料庫 SSL 設定
	if cfg.Database.SSLMode == "disable" {
		log.Fatal("Database SSL must be enabled in production")
	}

	// 檢查是否使用預設資料庫設定
	if cfg.Database.Host == "localhost" || cfg.Database.Host == "127.0.0.1" {
		log.Fatal("Production database must not use localhost")
	}
}
