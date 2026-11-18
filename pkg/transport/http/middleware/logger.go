package middleware

import (
	"log/slog"
	"time"

	"github.com/PavelLymin/diploma/pkg/logger"
	"github.com/gin-gonic/gin"
)

func Logger(log *slog.Logger) gin.HandlerFunc {
	if log == nil {
		log = slog.Default()
	}

	return func(c *gin.Context) {
		startTime := time.Now()

		requestID := c.GetHeader("X-Request-ID")

		scopedLog := log.With(slog.String("request_id", requestID))

		ctx := logger.WithLogger(
			c.Request.Context(),
			scopedLog,
		)

		c.Request = c.Request.WithContext(ctx)

		c.Next()

		statusCode := c.Writer.Status()
		latency := time.Since(startTime)

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}

		attrs := []slog.Attr{
			slog.Int("status", statusCode),
			slog.String("method", c.Request.Method),
			slog.String("route", route),
			slog.String("ip", c.ClientIP()),
			slog.Duration("latency", latency),
		}

		scopedLog.LogAttrs(
			c.Request.Context(),
			level(statusCode, len(c.Errors)),
			"request completed",
			attrs...,
		)
	}
}

func level(status int, errorsCount int) slog.Level {
	if status >= 500 || errorsCount > 0 {
		return slog.LevelError
	}
	if status >= 400 {
		return slog.LevelWarn
	}
	return slog.LevelInfo
}
