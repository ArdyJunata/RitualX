package middleware

import (
	"log/slog"
	"regexp"
	"strings"
)

const maxBodySize = 10240

var passwordRegex = regexp.MustCompile(`"password"\s*:\s*"[^"]*"`)

func sanitizeHeaders(headers map[string][]string) map[string][]string {
	sanitized := make(map[string][]string, len(headers))
	for k, v := range headers {
		if strings.EqualFold(k, "Authorization") {
			sanitized[k] = []string{"Bearer [REDACTED]"}
		} else {
			sanitized[k] = append([]string{}, v...)
		}
	}
	return sanitized
}

func captureBody(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	if len(body) > maxBodySize {
		return string(body[:maxBodySize]) + "...[truncated]"
	}
	return string(body)
}

func isMultipart(contentType string) bool {
	return strings.HasPrefix(contentType, "multipart/form-data")
}

func redactPasswords(body string) string {
	return passwordRegex.ReplaceAllString(body, `"password":"[REDACTED]"`)
}

func logLevel(status int) slog.Level {
	switch {
	case status >= 500:
		return slog.LevelError
	case status >= 400:
		return slog.LevelWarn
	default:
		return slog.LevelInfo
	}
}
