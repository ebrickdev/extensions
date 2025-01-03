package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/ebrickdev/ebrick/web"
	"golang.org/x/time/rate"
)

// RateLimiterConfig defines the configuration for the rate limiter.
type RateLimiterConfig struct {
	Requests int
	Window   time.Duration
	Burst    int
}

// RateLimiter manages rate limiting for clients.
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.Mutex
	config   RateLimiterConfig
}

// NewRateLimiter creates a new RateLimiter.
func NewRateLimiter(config RateLimiterConfig) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		config:   config,
	}
}

// getLimiter retrieves or creates a rate limiter for a given client.
func (rl *RateLimiter) getLimiter(client string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[client]
	if !exists {
		limiter = rate.NewLimiter(rate.Every(rl.config.Window/time.Duration(rl.config.Requests)), rl.config.Burst)
		rl.limiters[client] = limiter
	}
	return limiter
}

// Middleware enforces rate limiting based on client IP.
func (rl *RateLimiter) Middleware() web.HandlerFunc {
	return func(ctx web.Context) {
		clientIP := ctx.ClientIP()
		limiter := rl.getLimiter(clientIP)

		if !limiter.Allow() {
			ctx.JSON(http.StatusTooManyRequests, map[string]string{"error": "Rate limit exceeded"})
			return
		}

		ctx.Next()
	}
}
