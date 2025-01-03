package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/ebrickdev/ebrick/web"
)

type rateLimiter struct {
	requests map[string]int
	mu       sync.Mutex
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		requests: make(map[string]int),
		limit:    limit,
		window:   window,
	}
}

func (rl *rateLimiter) Middleware() web.HandlerFunc {
	return func(ctx web.Context) {
		clientIP := ctx.ClientIP()

		rl.mu.Lock()
		defer rl.mu.Unlock()

		// Reset count after the window duration
		if rl.requests[clientIP] == 0 {
			go func() {
				time.Sleep(rl.window)
				rl.mu.Lock()
				delete(rl.requests, clientIP)
				rl.mu.Unlock()
			}()
		}

		rl.requests[clientIP]++
		if rl.requests[clientIP] > rl.limit {
			ctx.JSON(http.StatusTooManyRequests, map[string]string{"error": "Rate limit exceeded"})
			return
		}

		// Process the next handler
		if nextFunc, ok := ctx.(interface{ Next() }); ok {
			nextFunc.Next()
		}
	}
}
