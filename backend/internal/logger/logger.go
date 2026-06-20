package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

type ctxKey string

const traceIDKey ctxKey = "trace_id"

var defaultLogger *slog.Logger

func Init(level string) {
	parsedLevel := parseLevel(level)
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: parsedLevel,
	})
	defaultLogger = slog.New(handler)
}

func Get() *slog.Logger {
	if defaultLogger == nil {
		return slog.Default()
	}
	return defaultLogger
}

func FromContext(ctx context.Context) *slog.Logger {
	l := Get()
	if traceID, ok := ctx.Value(traceIDKey).(string); ok && traceID != "" {
		return l.With("trace_id", traceID)
	}
	return l
}

func WithTraceID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, traceIDKey, id)
}

// SetLogger replaces the default logger (used for testing).
func SetLogger(l *slog.Logger) {
	defaultLogger = l
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
