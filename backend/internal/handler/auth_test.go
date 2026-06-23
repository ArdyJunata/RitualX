package handler

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/ArdyJunata/RitualX/backend/internal/model"
	"github.com/ArdyJunata/RitualX/backend/internal/repository"
	"github.com/ArdyJunata/RitualX/backend/internal/service"
)

func setupAuthTest(t *testing.T) (*fiber.App, *gorm.DB) {
	t.Helper()
	db := setupTestDB(t)

	app := fiber.New()
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	authService := service.NewAuthService(userRepo, refreshTokenRepo, "test_secret")
	app.Post("/api/v1/auth/register", Register(authService))
	app.Post("/api/v1/auth/login", Login(authService))
	app.Post("/api/v1/auth/refresh", Refresh(authService))
	app.Post("/api/v1/auth/logout", Logout(authService))

	return app, db
}

func cleanupUser(t *testing.T, db *gorm.DB, email string) {
	t.Helper()
	db.Where("email = ?", email).Delete(&model.User{})
}

func TestRegisterHandler_Success(t *testing.T) {
	app, db := setupAuthTest(t)
	defer cleanupUser(t, db, "handler-ok@test.com")

	body := `{"email":"handler-ok@test.com","username":"handlerok","password":"password123"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != 201 {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, want 201, body: %s", resp.StatusCode, respBody)
	}

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	if result["success"] != true {
		t.Errorf("success = %v, want true", result["success"])
	}

	data := result["data"].(map[string]interface{})
	if data["access_token"] == nil || data["access_token"] == "" {
		t.Error("access_token missing")
	}
	if data["refresh_token"] == nil || data["refresh_token"] == "" {
		t.Error("refresh_token missing")
	}

	user := data["user"].(map[string]interface{})
	if user["email"] != "handler-ok@test.com" {
		t.Errorf("email = %v, want handler-ok@test.com", user["email"])
	}
	if _, exists := user["password_hash"]; exists {
		t.Error("password_hash leaked in response")
	}
}

func TestRegisterHandler_ValidationError(t *testing.T) {
	app, _ := setupAuthTest(t)

	body := `{"email":"bad","username":"ab","password":"short"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != 400 {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	if result["success"] != false {
		t.Errorf("success = %v, want false", result["success"])
	}

	errObj := result["error"].(map[string]interface{})
	if errObj["code"] != "VALIDATION_ERROR" {
		t.Errorf("code = %v, want VALIDATION_ERROR", errObj["code"])
	}
	if errObj["details"] == nil {
		t.Error("details missing from validation error")
	}
}

func TestRegisterHandler_DuplicateEmail(t *testing.T) {
	app, db := setupAuthTest(t)
	defer cleanupUser(t, db, "dup-handler@test.com")

	body := `{"email":"dup-handler@test.com","username":"duphand1","password":"password123"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	app.Test(req)

	body = `{"email":"dup-handler@test.com","username":"duphand2","password":"password123"}`
	req = httptest.NewRequest("POST", "/api/v1/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != 409 {
		t.Errorf("status = %d, want 409", resp.StatusCode)
	}

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	errObj := result["error"].(map[string]interface{})
	if errObj["code"] != "EMAIL_TAKEN" {
		t.Errorf("code = %v, want EMAIL_TAKEN", errObj["code"])
	}
}

func TestRegisterHandler_InvalidJSON(t *testing.T) {
	app, _ := setupAuthTest(t)

	req := httptest.NewRequest("POST", "/api/v1/auth/register", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != 400 {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

func TestLoginHandler_Success(t *testing.T) {
	app, db := setupAuthTest(t)
	defer cleanupUser(t, db, "login-handler@test.com")

	// Register first
	reqReg := httptest.NewRequest("POST", "/api/v1/auth/register", strings.NewReader(`{"email":"login-handler@test.com","username":"loginhandler","password":"password123"}`))
	reqReg.Header.Set("Content-Type", "application/json")
	_, _ = app.Test(reqReg)

	// Login
	reqLogin := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(`{"email":"login-handler@test.com","password":"password123"}`))
	reqLogin.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(reqLogin)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	// Check if set-cookie header contains refresh_token
	cookies := resp.Header.Values("Set-Cookie")
	foundCookie := false
	for _, c := range cookies {
		if strings.Contains(c, "refresh_token=") && strings.Contains(c, "HttpOnly") {
			foundCookie = true
			break
		}
	}
	if !foundCookie {
		t.Error("expected refresh_token cookie with HttpOnly flag")
	}
}

func TestRefreshHandler_Success(t *testing.T) {
	app, db := setupAuthTest(t)
	defer cleanupUser(t, db, "refresh-handler@test.com")

	// Register to get cookie
	reqReg := httptest.NewRequest("POST", "/api/v1/auth/register", strings.NewReader(`{"email":"refresh-handler@test.com","username":"refreshhandler","password":"password123"}`))
	reqReg.Header.Set("Content-Type", "application/json")
	respReg, _ := app.Test(reqReg)

	var cookieStr string
	for _, c := range respReg.Header.Values("Set-Cookie") {
		if strings.Contains(c, "refresh_token=") {
			cookieStr = strings.Split(c, ";")[0]
			break
		}
	}

	// Refresh
	reqRef := httptest.NewRequest("POST", "/api/v1/auth/refresh", nil)
	reqRef.Header.Set("Cookie", cookieStr)
	resp, err := app.Test(reqRef)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}

func TestLogoutHandler_Success(t *testing.T) {
	app, db := setupAuthTest(t)
	defer cleanupUser(t, db, "logout-handler@test.com")

	// Register to get cookie
	reqReg := httptest.NewRequest("POST", "/api/v1/auth/register", strings.NewReader(`{"email":"logout-handler@test.com","username":"logouthandler","password":"password123"}`))
	reqReg.Header.Set("Content-Type", "application/json")
	respReg, _ := app.Test(reqReg)

	var cookieStr string
	for _, c := range respReg.Header.Values("Set-Cookie") {
		if strings.Contains(c, "refresh_token=") {
			cookieStr = strings.Split(c, ";")[0]
			break
		}
	}

	// Logout
	reqOut := httptest.NewRequest("POST", "/api/v1/auth/logout", nil)
	reqOut.Header.Set("Cookie", cookieStr)
	resp, err := app.Test(reqOut)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	// Check if cookie is cleared
	cookies := resp.Header.Values("Set-Cookie")
	foundClear := false
	for _, c := range cookies {
		if strings.Contains(c, "refresh_token=") && strings.Contains(c, "Max-Age=0") {
			foundClear = true
			break
		}
	}
	if !foundClear {
		t.Error("expected refresh_token cookie to be cleared")
	}
}
