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
}

type JWTConfig struct {
	Secret                 string
	AccessTokenExpiration  time.Duration
	RefreshTokenExpiration time.Duration
}

var AppConfig *Config

func LoadConfig() *Config {
	AppConfig = &Config{
		Env:      getEnv("APP_ENV", "development"),
		Port:     getEnv("PORT", "8080"),
		LogLevel: getEnv("LOG_LEVEL", "debug"),
		JWT: JWTConfig{
			Secret:                 getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
			AccessTokenExpiration:  getDurationEnv("JWT_ACCESS_TOKEN_EXPIRATION", 15*time.Minute),
			RefreshTokenExpiration: getDurationEnv("JWT_REFRESH_TOKEN_EXPIRATION", 7*24*time.Hour),
		},
	}
	return AppConfig
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
