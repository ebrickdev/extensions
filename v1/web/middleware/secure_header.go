package middleware

import (
	"github.com/ebrickdev/ebrick/web"
)

// SecureHeadersMiddleware adds security-related headers to HTTP responses.
func SecureHeadersMiddleware() web.HandlerFunc {
	return func(ctx web.Context) {
		ctx.SetHeader("Content-Security-Policy", "default-src 'self'")
		ctx.SetHeader("X-Content-Type-Options", "nosniff")
		ctx.SetHeader("X-Frame-Options", "DENY")
		ctx.SetHeader("X-XSS-Protection", "1; mode=block")
		ctx.Next()
	}
}
