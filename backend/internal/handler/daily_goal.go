package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/ArdyJunata/RitualX/backend/internal/service"
)

// GetDailyGoal returns today's daily goal summary (computed live).
func GetDailyGoal(dailyGoalService *service.DailyGoalService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := parseUserID(c)
		if err != nil {
			return err
		}

		goal, svcErr := dailyGoalService.GetToday(userID)
		if svcErr != nil {
			return handleServiceError(c, svcErr)
		}

		return success(c, fiber.StatusOK, goal)
	}
}

// GetDailyGoalHistory returns daily goal history for [from, to].
// Both query params are optional (YYYY-MM-DD); default is the last 30 days.
func GetDailyGoalHistory(dailyGoalService *service.DailyGoalService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := parseUserID(c)
		if err != nil {
			return err
		}

		to := time.Now().UTC()
		from := to.AddDate(0, 0, -29) // 30-day inclusive window

		if v := c.Query("to"); v != "" {
			parsed, perr := time.Parse("2006-01-02", v)
			if perr != nil {
				return errorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "to must be in YYYY-MM-DD format")
			}
			to = parsed
		}
		if v := c.Query("from"); v != "" {
			parsed, perr := time.Parse("2006-01-02", v)
			if perr != nil {
				return errorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "from must be in YYYY-MM-DD format")
			}
			from = parsed
		}

		goals, svcErr := dailyGoalService.GetHistory(userID, from, to)
		if svcErr != nil {
			return handleServiceError(c, svcErr)
		}

		return success(c, fiber.StatusOK, goals)
	}
}
