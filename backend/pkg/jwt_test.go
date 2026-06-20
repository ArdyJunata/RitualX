package pkg

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret-key"

func TestGenerateAccessToken_Valid(t *testing.T) {
	token, err := GenerateAccessToken("user-123", testSecret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("token is empty")
	}

	claims := &Claims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(testSecret), nil
	})
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}
	if !parsed.Valid {
		t.Error("token is not valid")
	}
	if claims.UserID != "user-123" {
		t.Errorf("UserID = %q, want %q", claims.UserID, "user-123")
	}
	if claims.Type != "access" {
		t.Errorf("Type = %q, want %q", claims.Type, "access")
	}

	exp := claims.ExpiresAt.Time
	expectedExp := time.Now().Add(15 * time.Minute)
	diff := expectedExp.Sub(exp)
	if diff < -5*time.Second || diff > 5*time.Second {
		t.Errorf("expiry off by %v, want ~15 min from now", diff)
	}
}

func TestGenerateRefreshToken_Valid(t *testing.T) {
	token, err := GenerateRefreshToken("user-456", testSecret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("token is empty")
	}

	claims := &Claims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(testSecret), nil
	})
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}
	if !parsed.Valid {
		t.Error("token is not valid")
	}
	if claims.UserID != "user-456" {
		t.Errorf("UserID = %q, want %q", claims.UserID, "user-456")
	}
	if claims.Type != "refresh" {
		t.Errorf("Type = %q, want %q", claims.Type, "refresh")
	}

	exp := claims.ExpiresAt.Time
	expectedExp := time.Now().Add(7 * 24 * time.Hour)
	diff := expectedExp.Sub(exp)
	if diff < -5*time.Second || diff > 5*time.Second {
		t.Errorf("expiry off by %v, want ~7 days from now", diff)
	}
}

func TestGenerateAccessToken_DifferentFromRefresh(t *testing.T) {
	access, _ := GenerateAccessToken("user-1", testSecret)
	refresh, _ := GenerateRefreshToken("user-1", testSecret)
	if access == refresh {
		t.Error("access and refresh tokens should be different")
	}
}
