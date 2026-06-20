package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/ArdyJunata/RitualX/backend/internal/logger"
)

func Trace() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Path() == "/api/v1/health" {
			return c.Next()
		}

		traceID := uuid.New().String()
		c.Locals("trace_id", traceID)

		method := c.Method()
		path := c.Path()
		host := c.Hostname()
		userAgent := c.Get("User-Agent")
		clientIP := c.IP()
		headers := sanitizeHeaders(c.GetReqHeaders())

		var reqBody string
		if isMultipart(c.Get("Content-Type")) {
			reqBody = "[multipart/form-data]"
		} else {
			reqBody = captureBody(c.Body())
		}

		start := time.Now()

		chainErr := c.Next()

		durationMs := time.Since(start).Milliseconds()
		status := c.Response().StatusCode()
		respBody := captureBody(c.Response().Body())

		c.Set("X-Trace-ID", traceID)

		if strings.HasPrefix(path, "/api/v1/auth/") {
			reqBody = redactPasswords(reqBody)
		}

		var errMsg string
		if chainErr != nil {
			errMsg = chainErr.Error()
		}

		level := logLevel(status)
		log := logger.Get()

		log.Log(
			c.UserContext(),
			level,
			"request completed",
			"trace_id", traceID,
			"method", method,
			"path", path,
			"host", host,
			"user_agent", userAgent,
			"request_headers", headers,
			"request_body", reqBody,
			"response_status", status,
			"response_body", respBody,
			"duration_ms", durationMs,
			"client_ip", clientIP,
			"error", errMsg,
		)

		return chainErr
	}
}
