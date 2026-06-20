package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/ArdyJunata/RitualX/backend/internal/service"
)

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

		return success(c, fiber.StatusCreated, resp)
	}
}
