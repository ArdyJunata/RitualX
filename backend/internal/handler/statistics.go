package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/ArdyJunata/RitualX/backend/internal/service"
)

// GetStatsOverview returns overall stats for the authenticated user.
func GetStatsOverview(statisticsService *service.StatisticsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := parseUserID(c)
		if err != nil {
			return err
		}
		data, svcErr := statisticsService.GetOverview(userID)
		if svcErr != nil {
			return handleServiceError(c, svcErr)
		}
		return success(c, fiber.StatusOK, data)
	}
}

// GetRoutineStatistics returns per-routine statistics.
func GetRoutineStatistics(statisticsService *service.StatisticsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := parseUserID(c)
		if err != nil {
			return err
		}
		routineID, err := parseParamID(c, "id")
		if err != nil {
			return err
		}
		data, svcErr := statisticsService.GetRoutineStats(userID, routineID)
		if svcErr != nil {
			return handleServiceError(c, svcErr)
		}
		return success(c, fiber.StatusOK, data)
	}
}

// GetWeeklySummary returns this-week vs last-week completion totals.
func GetWeeklySummary(statisticsService *service.StatisticsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := parseUserID(c)
		if err != nil {
			return err
		}
		data, svcErr := statisticsService.GetWeeklySummary(userID)
		if svcErr != nil {
			return handleServiceError(c, svcErr)
		}
		return success(c, fiber.StatusOK, data)
	}
}

// GetMonthlySummary returns this-month vs last-month completion totals.
func GetMonthlySummary(statisticsService *service.StatisticsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := parseUserID(c)
		if err != nil {
			return err
		}
		data, svcErr := statisticsService.GetMonthlySummary(userID)
		if svcErr != nil {
			return handleServiceError(c, svcErr)
		}
		return success(c, fiber.StatusOK, data)
	}
}
