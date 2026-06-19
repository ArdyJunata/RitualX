package config

import (
	"os"
	"testing"
)

func TestLoad_WithAllEnvVars(t *testing.T) {
	os.Setenv("APP_PORT", "9090")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("JWT_SECRET", "testsecret")
	os.Setenv("LOG_LEVEL", "debug")
	defer func() {
		os.Unsetenv("APP_PORT")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("LOG_LEVEL")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.AppPort != "9090" {
		t.Errorf("AppPort = %q, want %q", cfg.AppPort, "9090")
	}
	if cfg.DBHost != "localhost" {
		t.Errorf("DBHost = %q, want %q", cfg.DBHost, "localhost")
	}
	if cfg.DBPort != "5433" {
		t.Errorf("DBPort = %q, want %q", cfg.DBPort, "5433")
	}
	if cfg.DBUser != "testuser" {
		t.Errorf("DBUser = %q, want %q", cfg.DBUser, "testuser")
	}
	if cfg.DBPassword != "testpass" {
		t.Errorf("DBPassword = %q, want %q", cfg.DBPassword, "testpass")
	}
	if cfg.DBName != "testdb" {
		t.Errorf("DBName = %q, want %q", cfg.DBName, "testdb")
	}
	if cfg.JWTSecret != "testsecret" {
		t.Errorf("JWTSecret = %q, want %q", cfg.JWTSecret, "testsecret")
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, "debug")
	}
}

func TestLoad_Defaults(t *testing.T) {
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("JWT_SECRET", "testsecret")
	os.Unsetenv("APP_PORT")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("LOG_LEVEL")
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("JWT_SECRET")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.AppPort != "8080" {
		t.Errorf("AppPort = %q, want default %q", cfg.AppPort, "8080")
	}
	if cfg.DBPort != "5432" {
		t.Errorf("DBPort = %q, want default %q", cfg.DBPort, "5432")
	}
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %q, want default %q", cfg.LogLevel, "info")
	}
}

func TestLoad_MissingRequired(t *testing.T) {
	os.Clearenv()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing required vars, got nil")
	}
}
