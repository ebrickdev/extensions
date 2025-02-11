package middleware

import (
	"time"

	"github.com/ebrickdev/ebrick/logger"
	"github.com/ebrickdev/ebrick/transport/http"
)

// RequestLoggingMiddleware logs details for each HTTP request.
func RequestLoggingMiddleware(log logger.Logger) http.HandlerFunc {
	return func(ctx *http.Context) {
		start := time.Now()

		// Ensure latency is always measured, even if an error occurs
		defer func() {
			latency := time.Since(start)

			// Log the response details
			log.Info("HTTP Request",
				logger.String("method", ctx.Request.Method),
				logger.String("path", ctx.Request.URL.Path),
				logger.String("query", ctx.Request.URL.RawQuery),
				logger.String("client_ip", ctx.ClientIP()),
				logger.String("user_agent", ctx.Request.UserAgent()),
				logger.Int("status_code", ctx.Writer.Status()),
				logger.String("latency", latency.String()),
			)
		}()

		// Process the request
		ctx.Next()
	}
}
