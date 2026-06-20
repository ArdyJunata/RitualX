package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/ArdyJunata/RitualX/backend/internal/service"
)

func success(c *fiber.Ctx, status int, data interface{}) error {
	return c.Status(status).JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}

func errorResponse(c *fiber.Ctx, status int, code string, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"success": false,
		"error": fiber.Map{
			"code":    code,
			"message": message,
		},
	})
}

func handleServiceError(c *fiber.Ctx, err error) error {
	svcErr, ok := err.(*service.ServiceError)
	if !ok {
		return errorResponse(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "unexpected error")
	}

	status := mapErrorCodeToStatus(svcErr.Code)

	errResp := fiber.Map{
		"code":    svcErr.Code,
		"message": svcErr.Message,
	}
	if len(svcErr.Details) > 0 {
		errResp["details"] = svcErr.Details
	}

	return c.Status(status).JSON(fiber.Map{
		"success": false,
		"error":   errResp,
	})
}

func mapErrorCodeToStatus(code string) int {
	switch code {
	case "VALIDATION_ERROR", "INVALID_REQUEST":
		return fiber.StatusBadRequest
	case "EMAIL_TAKEN", "USERNAME_TAKEN":
		return fiber.StatusConflict
	default:
		return fiber.StatusInternalServerError
	}
}
