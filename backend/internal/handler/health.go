package handler

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func HealthCheck(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var result int
		if err := db.Raw("SELECT 1").Scan(&result).Error; err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "DB_UNHEALTHY",
					"message": "database connection failed",
				},
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"data": fiber.Map{
				"status":  "healthy",
				"version": "0.1.0",
			},
		})
	}
}
