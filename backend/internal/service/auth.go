package service

import (
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/ArdyJunata/RitualX/backend/internal/model"
	"github.com/ArdyJunata/RitualX/backend/internal/repository"
	"github.com/ArdyJunata/RitualX/backend/pkg"
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`)

type RegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	User         model.User `json:"user"`
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"-"`
}

type RegisterResponse struct {
	User         model.User `json:"user"`
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"-"`
}

type AuthService struct {
	userRepo         *repository.UserRepository
	refreshTokenRepo *repository.RefreshTokenRepository
	jwtSecret        string
}

func NewAuthService(userRepo *repository.UserRepository, rtRepo *repository.RefreshTokenRepository, jwtSecret string) *AuthService {
	return &AuthService{userRepo: userRepo, refreshTokenRepo: rtRepo, jwtSecret: jwtSecret}
}

func (s *AuthService) Register(req RegisterRequest, ip, userAgent string) (*RegisterResponse, error) {
	if errs := validateRegister(req); len(errs) > 0 {
		return nil, &ServiceError{
			Code:    "VALIDATION_ERROR",
			Message: "Validation failed",
			Details: errs,
		}
	}

	existing, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "internal error"}
	}
	if existing != nil {
		return nil, &ServiceError{Code: "EMAIL_TAKEN", Message: "email already registered"}
	}

	existing, err = s.userRepo.FindByUsername(req.Username)
	if err != nil {
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "internal error"}
	}
	if existing != nil {
		return nil, &ServiceError{Code: "USERNAME_TAKEN", Message: "username already taken"}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "internal error"}
	}

	user := &model.User{
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: string(hash),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, mapCreateError(err)
	}

	accessToken, err := pkg.GenerateAccessToken(user.ID.String(), s.jwtSecret)
	if err != nil {
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "internal error"}
	}

	refreshToken, err := pkg.GenerateRefreshToken(user.ID.String(), s.jwtSecret)
	if err != nil {
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "internal error"}
	}

	rt := &model.RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		IPAddress: ip,
		UserAgent: userAgent,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	_ = s.refreshTokenRepo.Create(rt)

	return &RegisterResponse{
		User:         *user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func mapCreateError(err error) *ServiceError {
	msg := err.Error()
	if strings.Contains(msg, "duplicate key") || strings.Contains(msg, "unique constraint") {
		if strings.Contains(msg, "email") {
			return &ServiceError{Code: "EMAIL_TAKEN", Message: "email already registered"}
		}
		if strings.Contains(msg, "username") {
			return &ServiceError{Code: "USERNAME_TAKEN", Message: "username already taken"}
		}
	}
	return &ServiceError{Code: "INTERNAL_ERROR", Message: "internal error"}
}

func validateRegister(req RegisterRequest) []FieldError {
	var errs []FieldError

	if !isValidEmail(req.Email) {
		errs = append(errs, FieldError{Field: "email", Message: "invalid email format"})
	}
	if !usernameRegex.MatchString(req.Username) {
		errs = append(errs, FieldError{Field: "username", Message: "must be 3-20 characters, alphanumeric and underscore only"})
	}
	if len(req.Password) < 8 {
		errs = append(errs, FieldError{Field: "password", Message: "must be at least 8 characters"})
	}

	return errs
}

func isValidEmail(email string) bool {
	if email == "" {
		return false
	}
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return false
	}
	return strings.Contains(parts[1], ".")
}

func (s *AuthService) Login(req LoginRequest, ip, userAgent string) (*LoginResponse, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "internal error"}
	}
	if user == nil {
		return nil, &ServiceError{Code: "INVALID_CREDENTIALS", Message: "invalid email or password"}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, &ServiceError{Code: "INVALID_CREDENTIALS", Message: "invalid email or password"}
	}

	accessToken, err := pkg.GenerateAccessToken(user.ID.String(), s.jwtSecret)
	if err != nil {
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "could not generate access token"}
	}
	refreshTokenStr, err := pkg.GenerateRefreshToken(user.ID.String(), s.jwtSecret)
	if err != nil {
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "could not generate refresh token"}
	}

	rt := &model.RefreshToken{
		UserID:    user.ID,
		Token:     refreshTokenStr,
		IPAddress: ip,
		UserAgent: userAgent,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := s.refreshTokenRepo.Create(rt); err != nil {
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "could not save session"}
	}

	return &LoginResponse{
		User:         *user,
		AccessToken:  accessToken,
		RefreshToken: refreshTokenStr,
	}, nil
}

func (s *AuthService) Refresh(tokenStr string) (string, error) {
	if tokenStr == "" {
		return "", &ServiceError{Code: "UNAUTHORIZED", Message: "missing refresh token"}
	}
	rt, err := s.refreshTokenRepo.FindByToken(tokenStr)
	if err != nil || rt == nil {
		return "", &ServiceError{Code: "UNAUTHORIZED", Message: "invalid refresh token"}
	}
	if time.Now().After(rt.ExpiresAt) {
		_ = s.refreshTokenRepo.Delete(tokenStr)
		return "", &ServiceError{Code: "UNAUTHORIZED", Message: "refresh token expired"}
	}

	accessToken, err := pkg.GenerateAccessToken(rt.UserID.String(), s.jwtSecret)
	if err != nil {
		return "", &ServiceError{Code: "INTERNAL_ERROR", Message: "could not generate access token"}
	}
	return accessToken, nil
}

func (s *AuthService) Logout(tokenStr string) error {
	if tokenStr == "" {
		return nil
	}
	return s.refreshTokenRepo.Delete(tokenStr)
}
