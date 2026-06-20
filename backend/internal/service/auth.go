package service

import (
	"regexp"
	"strings"

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

type RegisterResponse struct {
	User         model.User `json:"user"`
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token"`
}

type AuthService struct {
	userRepo  *repository.UserRepository
	jwtSecret string
}

func NewAuthService(userRepo *repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{userRepo: userRepo, jwtSecret: jwtSecret}
}

func (s *AuthService) Register(req RegisterRequest) (*RegisterResponse, error) {
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
