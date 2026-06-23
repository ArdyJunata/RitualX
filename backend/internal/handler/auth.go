package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/ArdyJunata/RitualX/backend/internal/service"
)

func setRefreshTokenCookie(c *fiber.Ctx, token string) {
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60,
		Secure:   true,
		HTTPOnly: true,
		SameSite: "Strict",
	})
}

func clearRefreshTokenCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   true,
		HTTPOnly: true,
		SameSite: "Strict",
	})
}

func Register(authService *service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req service.RegisterRequest
		if err := c.BodyParser(&req); err != nil {
			return errorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		}

		resp, err := authService.Register(req)
		if err != nil {
			return handleServiceError(c, err)
		}

		setRefreshTokenCookie(c, resp.RefreshToken)
		return success(c, fiber.StatusCreated, resp)
	}
}

func Login(authService *service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req service.LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return errorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		}

		resp, err := authService.Login(req, c.IP(), string(c.Request().Header.UserAgent()))
		if err != nil {
			return handleServiceError(c, err)
		}

		setRefreshTokenCookie(c, resp.RefreshToken)
		return success(c, fiber.StatusOK, resp)
	}
}

func Refresh(authService *service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Cookies("refresh_token")
		newAccessToken, err := authService.Refresh(token)
		if err != nil {
			return handleServiceError(c, err)
		}
		return success(c, fiber.StatusOK, fiber.Map{"access_token": newAccessToken})
	}
}

func Logout(authService *service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Cookies("refresh_token")
		_ = authService.Logout(token)
		clearRefreshTokenCookie(c)
		return success(c, fiber.StatusOK, fiber.Map{"message": "logged out"})
	}
}
