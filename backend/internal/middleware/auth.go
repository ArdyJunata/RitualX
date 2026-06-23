package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/ArdyJunata/RitualX/backend/pkg"
)

func RequireAuth(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "UNAUTHORIZED",
					"message": "missing or invalid authorization header",
				},
			})
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := pkg.ValidateToken(tokenStr, jwtSecret)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "UNAUTHORIZED",
					"message": "invalid or expired token",
				},
			})
		}

		c.Locals("user_id", claims.UserID)
		return c.Next()
	}
}
