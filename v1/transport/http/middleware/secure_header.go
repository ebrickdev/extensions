package middleware

import (
	"github.com/ebrickdev/ebrick/transport/http"
)

// SecureHeadersMiddleware adds security-related headers to HTTP responses.
func SecureHeadersMiddleware() http.HandlerFunc {
	return func(ctx *http.Context) {
		ctx.Header("Content-Security-Policy", "default-src 'self'")
		ctx.Header("X-Content-Type-Options", "nosniff")
		ctx.Header("X-Frame-Options", "DENY")
		ctx.Header("X-XSS-Protection", "1; mode=block")
		ctx.Next()
	}
}
