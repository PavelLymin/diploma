package logger

import (
	"context"
	"log/slog"
)

type ctxLoggerKey struct{}

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxLoggerKey{}, logger)
}

func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(ctxLoggerKey{}).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}
