package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort    string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	JWTSecret  string
	LogLevel   string
}

func Load() (*Config, error) {
	_ = godotenv.Load() // non-fatal if .env missing

	cfg := &Config{
		AppPort:    getEnvOrDefault("APP_PORT", "8080"),
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     getEnvOrDefault("DB_PORT", "5432"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		JWTSecret:  os.Getenv("JWT_SECRET"),
		LogLevel:   getEnvOrDefault("LOG_LEVEL", "info"),
	}

	if cfg.DBHost == "" {
		return nil, fmt.Errorf("required env var DB_HOST is not set")
	}
	if cfg.DBUser == "" {
		return nil, fmt.Errorf("required env var DB_USER is not set")
	}
	if cfg.DBPassword == "" {
		return nil, fmt.Errorf("required env var DB_PASSWORD is not set")
	}
	if cfg.DBName == "" {
		return nil, fmt.Errorf("required env var DB_NAME is not set")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("required env var JWT_SECRET is not set")
	}

	return cfg, nil
}

func getEnvOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
