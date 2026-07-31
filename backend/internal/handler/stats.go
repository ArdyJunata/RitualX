package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/ArdyJunata/RitualX/backend/internal/service"
)

// GetRoutineHeatmap returns a date->count map for the routine within a year (default current year).
func GetRoutineHeatmap(statsService *service.StatsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := parseUserID(c)
		if err != nil {
			return err
		}

		routineID, err := parseParamID(c, "id")
		if err != nil {
			return err
		}

		year := time.Now().UTC().Year()
		if v := c.Query("year"); v != "" {
			parsed, perr := strconv.Atoi(v)
			if perr != nil || parsed < 2000 || parsed > 2100 {
				return errorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "year must be a valid year between 2000 and 2100")
			}
			year = parsed
		}

		data, svcErr := statsService.GetHeatmap(userID, routineID, year)
		if svcErr != nil {
			return handleServiceError(c, svcErr)
		}

		return success(c, fiber.StatusOK, data)
	}
}

// GetRoutineProgress returns current-period progress (completed/target/period) for the routine.
func GetRoutineProgress(statsService *service.StatsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := parseUserID(c)
		if err != nil {
			return err
		}

		routineID, err := parseParamID(c, "id")
		if err != nil {
			return err
		}

		data, svcErr := statsService.GetProgress(userID, routineID)
		if svcErr != nil {
			return handleServiceError(c, svcErr)
		}

		return success(c, fiber.StatusOK, data)
	}
}
