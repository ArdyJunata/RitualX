package pkg

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"user_id"`
	Type   string `json:"type"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(userID string, secret string) (string, error) {
	return generateToken(userID, secret, "access", 15*time.Minute)
}

func GenerateRefreshToken(userID string, secret string) (string, error) {
	return generateToken(userID, secret, "refresh", 7*24*time.Hour)
}

func generateToken(userID, secret, tokenType string, expiry time.Duration) (string, error) {
	claims := &Claims{
		UserID: userID,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
