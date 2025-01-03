package middleware

import (
	"time"

	"github.com/ebrickdev/ebrick/logger"
	"github.com/ebrickdev/ebrick/web"
)

// LoggerMiddleware logs request details for each HTTP request
func RequestLoggingMiddleware(log logger.Logger) web.HandlerFunc {
	return func(ctx web.Context) {
		start := time.Now()

		// Process the request
		if nextFunc, ok := ctx.(interface{ Next() }); ok {
			nextFunc.Next()
		}

		// Log details
		log.Info("HTTP Request",
			logger.String("method", ctx.Request().Method),
			logger.String("path", ctx.Request().URL.Path),
			logger.String("query", ctx.Request().URL.RawQuery),
			logger.String("client_ip", ctx.ClientIP()),
			logger.String("user_agent", ctx.Request().UserAgent()),
			logger.Any("latency", time.Since(start)),
		)
	}
}
