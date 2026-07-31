package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/ArdyJunata/RitualX/backend/internal/service"
)

// GetRoutineStreak returns the persisted streak for a single routine.
func GetRoutineStreak(streakService *service.StreakService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := parseUserID(c)
		if err != nil {
			return err
		}

		routineID, err := parseParamID(c, "id")
		if err != nil {
			return err
		}

		streak, svcErr := streakService.GetByRoutine(userID, routineID)
		if svcErr != nil {
			return handleServiceError(c, svcErr)
		}

		return success(c, fiber.StatusOK, streak)
	}
}

// ListStreaks returns all streaks for the authenticated user.
func ListStreaks(streakService *service.StreakService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := parseUserID(c)
		if err != nil {
			return err
		}

		streaks, svcErr := streakService.ListByUser(userID)
		if svcErr != nil {
			return handleServiceError(c, svcErr)
		}

		return success(c, fiber.StatusOK, streaks)
	}
}
