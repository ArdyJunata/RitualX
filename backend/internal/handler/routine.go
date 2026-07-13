package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/ArdyJunata/RitualX/backend/internal/service"
)

func CreateRoutine(routineService *service.RoutineService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userIDStr, ok := c.Locals("user_id").(string)
		if !ok || userIDStr == "" {
			return errorResponse(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "missing user identity")
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return errorResponse(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "invalid user identity")
		}

		var req service.CreateRoutineRequest
		if err := c.BodyParser(&req); err != nil {
			return errorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		}

		routine, svcErr := routineService.Create(userID, req)
		if svcErr != nil {
			return handleServiceError(c, svcErr)
		}

		return success(c, fiber.StatusCreated, routine)
	}
}
