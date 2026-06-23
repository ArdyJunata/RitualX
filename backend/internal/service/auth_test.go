package service

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/ArdyJunata/RitualX/backend/internal/model"
	"github.com/ArdyJunata/RitualX/backend/internal/repository"
	"github.com/ArdyJunata/RitualX/backend/pkg"
)

const testJWTSecret = "test-jwt-secret"

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := "host=localhost port=5432 user=ritualx password=ritualx_dev dbname=ritualx sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skip("PostgreSQL not available, skipping integration test")
	}
	return db
}

func cleanupUser(t *testing.T, db *gorm.DB, email string) {
	t.Helper()
	db.Where("email = ?", email).Delete(&model.User{})
}

func newAuthService(t *testing.T) (*AuthService, *gorm.DB) {
	t.Helper()
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	svc := NewAuthService(userRepo, refreshTokenRepo, testJWTSecret)
	return svc, db
}

func TestRegister_Success(t *testing.T) {
	svc, db := newAuthService(t)
	defer cleanupUser(t, db, "register-ok@test.com")

	resp, err := svc.Register(RegisterRequest{
		Email:    "register-ok@test.com",
		Username: "registerok",
		Password: "password123",
	}, "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if resp.User.Email != "register-ok@test.com" {
		t.Errorf("email = %q, want %q", resp.User.Email, "register-ok@test.com")
	}
	if resp.User.Username != "registerok" {
		t.Errorf("username = %q, want %q", resp.User.Username, "registerok")
	}
	if resp.AccessToken == "" {
		t.Error("access token is empty")
	}
	if resp.RefreshToken == "" {
		t.Error("refresh token is empty")
	}

	claims := &pkg.Claims{}
	jwt.ParseWithClaims(resp.AccessToken, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(testJWTSecret), nil
	})
	if claims.UserID != resp.User.ID.String() {
		t.Errorf("token user_id = %q, want %q", claims.UserID, resp.User.ID.String())
	}
	if claims.Type != "access" {
		t.Errorf("token type = %q, want %q", claims.Type, "access")
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	svc, db := newAuthService(t)
	defer cleanupUser(t, db, "dup-email@test.com")

	req := RegisterRequest{
		Email:    "dup-email@test.com",
		Username: "dupuser1",
		Password: "password123",
	}
	_, _ = svc.Register(req, "127.0.0.1", "test-agent")

	req.Username = "dupuser2"
	_, err := svc.Register(req, "127.0.0.1", "test-agent")
	if err == nil {
		t.Fatal("expected error for duplicate email")
	}
	svcErr, ok := err.(*ServiceError)
	if !ok {
		t.Fatalf("expected *ServiceError, got %T", err)
	}
	if svcErr.Code != "EMAIL_TAKEN" {
		t.Errorf("code = %q, want %q", svcErr.Code, "EMAIL_TAKEN")
	}
}

func TestRegister_DuplicateUsername(t *testing.T) {
	svc, db := newAuthService(t)
	defer cleanupUser(t, db, "dup-uname1@test.com")
	defer cleanupUser(t, db, "dup-uname2@test.com")

	req := RegisterRequest{
		Email:    "dup-uname1@test.com",
		Username: "sameuser",
		Password: "password123",
	}
	_, _ = svc.Register(req, "127.0.0.1", "test-agent")

	req.Email = "dup-uname2@test.com"
	_, err := svc.Register(req, "127.0.0.1", "test-agent")
	if err == nil {
		t.Fatal("expected error for duplicate username")
	}
	svcErr, ok := err.(*ServiceError)
	if !ok {
		t.Fatalf("expected *ServiceError, got %T", err)
	}
	if svcErr.Code != "USERNAME_TAKEN" {
		t.Errorf("code = %q, want %q", svcErr.Code, "USERNAME_TAKEN")
	}
}

func TestRegister_ValidationErrors(t *testing.T) {
	svc, _ := newAuthService(t)

	tests := []struct {
		name  string
		req   RegisterRequest
		field string
	}{
		{"empty email", RegisterRequest{Email: "", Username: "valid_u", Password: "12345678"}, "email"},
		{"invalid email", RegisterRequest{Email: "notanemail", Username: "valid_u", Password: "12345678"}, "email"},
		{"short username", RegisterRequest{Email: "a@b.com", Username: "ab", Password: "12345678"}, "username"},
		{"long username", RegisterRequest{Email: "a@b.com", Username: "abcdefghijklmnopqrstu", Password: "12345678"}, "username"},
		{"invalid chars username", RegisterRequest{Email: "a@b.com", Username: "bad-user!", Password: "12345678"}, "username"},
		{"short password", RegisterRequest{Email: "a@b.com", Username: "valid_u", Password: "1234567"}, "password"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Register(tt.req, "127.0.0.1", "test-agent")
			if err == nil {
				t.Fatal("expected validation error")
			}
			svcErr, ok := err.(*ServiceError)
			if !ok {
				t.Fatalf("expected *ServiceError, got %T", err)
			}
			if svcErr.Code != "VALIDATION_ERROR" {
				t.Errorf("code = %q, want %q", svcErr.Code, "VALIDATION_ERROR")
			}
			found := false
			for _, d := range svcErr.Details {
				if d.Field == tt.field {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected field error for %q, got details: %v", tt.field, svcErr.Details)
			}
		})
	}
}

func TestRegister_PasswordHashed(t *testing.T) {
	svc, db := newAuthService(t)
	defer cleanupUser(t, db, "test@example.com")

	_, err := svc.Register(RegisterRequest{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "password123",
	}, "127.0.0.1", "test-agent")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var user model.User
	db.Where("email = ?", "test@example.com").First(&user)
	if user.PasswordHash == "password123" {
		t.Error("password stored as plaintext")
	}
	if user.PasswordHash == "" {
		t.Error("password_hash is empty")
	}
}

func TestAuthService_Login(t *testing.T) {
	svc, db := newAuthService(t)
	defer cleanupUser(t, db, "login@example.com")

	req := RegisterRequest{
		Email:    "login@example.com",
		Username: "loginuser",
		Password: "password123",
	}
	_, _ = svc.Register(req, "127.0.0.1", "test-agent")

	// Valid login
	resp, err := svc.Login(LoginRequest{
		Email:    "login@example.com",
		Password: "password123",
	}, "127.0.0.1", "test-agent")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.AccessToken == "" {
		t.Error("expected access token")
	}
	if resp.RefreshToken == "" {
		t.Error("expected refresh token")
	}

	// Invalid login
	_, err = svc.Login(LoginRequest{
		Email:    "login@example.com",
		Password: "wrongpassword",
	}, "127.0.0.1", "test-agent")

	if err == nil {
		t.Error("expected error for invalid password")
	}
}

func TestAuthService_Refresh(t *testing.T) {
	svc, db := newAuthService(t)
	defer cleanupUser(t, db, "refresh@example.com")

	req := RegisterRequest{
		Email:    "refresh@example.com",
		Username: "refreshuser",
		Password: "password123",
	}
	regResp, _ := svc.Register(req, "127.0.0.1", "test-agent")

	token, err := svc.Refresh(regResp.RefreshToken)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token == "" {
		t.Error("expected new access token")
	}

	_, err = svc.Refresh("invalid-token")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestAuthService_Logout(t *testing.T) {
	svc, db := newAuthService(t)
	defer cleanupUser(t, db, "logout@example.com")

	req := RegisterRequest{
		Email:    "logout@example.com",
		Username: "logoutuser",
		Password: "password123",
	}
	regResp, _ := svc.Register(req, "127.0.0.1", "test-agent")

	err := svc.Logout(regResp.RefreshToken)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Token should be invalid now
	_, err = svc.Refresh(regResp.RefreshToken)
	if err == nil {
		t.Error("expected error trying to refresh after logout")
	}
}
