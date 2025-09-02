package config

import "os"

type Config struct {
	Env      string
	Port     string
	LogLevel string
}

var AppConfig *Config

func LoadConfig() *Config {
	AppConfig = &Config{
		Env:      getEnv("APP_ENV", "development"),
		Port:     getEnv("PORT", "8080"),
		LogLevel: getEnv("LOG_LEVEL", "debug"),
	}
	return AppConfig
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
