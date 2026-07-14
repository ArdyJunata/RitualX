package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/ArdyJunata/RitualX/backend/internal/service"
)

func CreateRoutine(routineService *service.RoutineService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := parseUserID(c)
		if err != nil {
			return err
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

func ListRoutines(routineService *service.RoutineService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := parseUserID(c)
		if err != nil {
			return err
		}

		routines, svcErr := routineService.List(userID)
		if svcErr != nil {
			return handleServiceError(c, svcErr)
		}

		return success(c, fiber.StatusOK, routines)
	}
}

func GetRoutine(routineService *service.RoutineService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := parseUserID(c)
		if err != nil {
			return err
		}

		routineID, err := parseParamID(c, "id")
		if err != nil {
			return err
		}

		routine, svcErr := routineService.GetByID(userID, routineID)
		if svcErr != nil {
			return handleServiceError(c, svcErr)
		}

		return success(c, fiber.StatusOK, routine)
	}
}

func UpdateRoutine(routineService *service.RoutineService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := parseUserID(c)
		if err != nil {
			return err
		}

		routineID, err := parseParamID(c, "id")
		if err != nil {
			return err
		}

		var req service.UpdateRoutineRequest
		if err := c.BodyParser(&req); err != nil {
			return errorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		}

		routine, svcErr := routineService.Update(userID, routineID, req)
		if svcErr != nil {
			return handleServiceError(c, svcErr)
		}

		return success(c, fiber.StatusOK, routine)
	}
}

func DeleteRoutine(routineService *service.RoutineService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := parseUserID(c)
		if err != nil {
			return err
		}

		routineID, err := parseParamID(c, "id")
		if err != nil {
			return err
		}

		svcErr := routineService.Delete(userID, routineID)
		if svcErr != nil {
			return handleServiceError(c, svcErr)
		}

		return success(c, fiber.StatusOK, nil)
	}
}

func ReorderRoutines(routineService *service.RoutineService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := parseUserID(c)
		if err != nil {
			return err
		}

		var req service.ReorderRoutineRequest
		if err := c.BodyParser(&req); err != nil {
			return errorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		}

		svcErr := routineService.Reorder(userID, req)
		if svcErr != nil {
			return handleServiceError(c, svcErr)
		}

		return success(c, fiber.StatusOK, nil)
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// parseUserID extracts and parses the user_id from Fiber locals (set by RequireAuth middleware).
func parseUserID(c *fiber.Ctx) (uuid.UUID, error) {
	str, ok := c.Locals("user_id").(string)
	if !ok || str == "" {
		return uuid.UUID{}, errorResponse(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "missing user identity")
	}
	id, err := uuid.Parse(str)
	if err != nil {
		return uuid.UUID{}, errorResponse(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "invalid user identity")
	}
	return id, nil
}

// parseParamID parses a named URL param as a UUID.
func parseParamID(c *fiber.Ctx, param string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Params(param))
	if err != nil {
		return uuid.UUID{}, errorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "invalid "+param)
	}
	return id, nil
}
