package handler

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := "host=localhost port=5432 user=ritualx password=ritualx_dev dbname=ritualx sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skip("PostgreSQL not available, skipping integration test")
	}
	return db
}

func TestHealthCheck_Healthy(t *testing.T) {
	db := setupTestDB(t)

	app := fiber.New()
	app.Get("/api/v1/health", HealthCheck(db))

	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if result["success"] != true {
		t.Errorf("success = %v, want true", result["success"])
	}
	data := result["data"].(map[string]interface{})
	if data["status"] != "healthy" {
		t.Errorf("status = %v, want healthy", data["status"])
	}
	if data["version"] != "0.1.0" {
		t.Errorf("version = %v, want 0.1.0", data["version"])
	}
}

func TestHealthCheck_Unhealthy(t *testing.T) {
	dsn := "host=localhost port=9999 user=invalid password=invalid dbname=invalid sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skip("Cannot create gorm instance for unhealthy test")
	}

	app := fiber.New()
	app.Get("/api/v1/health", HealthCheck(db))

	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != 503 {
		t.Errorf("status = %d, want 503", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if result["success"] != false {
		t.Errorf("success = %v, want false", result["success"])
	}
}
