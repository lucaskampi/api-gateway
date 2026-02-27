package middleware

import (
	"context"
	"sync"

	"github.com/gofiber/fiber/v3"

	"api-gateway/internal/adapter/ratelimit"
)

var (
	limiters   = make(map[string]*ratelimit.TokenBucket)
	limitersMu sync.RWMutex
)

func RateLimit(rps, burst int) fiber.Handler {
	return func(c fiber.Ctx) error {
		path := c.Route().Path
		limitersMu.RLock()
		limiter, exists := limiters[path]
		limitersMu.RUnlock()

		if !exists {
			limitersMu.Lock()
			if limiter, exists = limiters[path]; !exists {
				limiter = ratelimit.NewTokenBucket(rps, burst)
				limiters[path] = limiter
			}
			limitersMu.Unlock()
		}

		key := c.IP()
		allowed, err := limiter.Allow(context.Background(), key)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "rate limit error",
			})
		}

		if !allowed {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       "rate limit exceeded",
				"retry_after": "1s",
			})
		}
		return c.Next()
	}
}
