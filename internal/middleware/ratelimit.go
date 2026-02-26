package middleware

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
)

type RateLimiter struct {
	tokens   map[string]*tokenBucket
	mu       sync.RWMutex
	rps      int
	burst    int
	refillMs int
}

type tokenBucket struct {
	tokens     float64
	lastRefill time.Time
}

func NewRateLimiter(rps, burst int) *RateLimiter {
	rl := &RateLimiter{
		tokens:   make(map[string]*tokenBucket),
		rps:      rps,
		burst:    burst,
		refillMs: 1000,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, bucket := range rl.tokens {
			if now.Sub(bucket.lastRefill) > 10*time.Minute {
				delete(rl.tokens, key)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	bucket, exists := rl.tokens[key]

	if !exists {
		rl.tokens[key] = &tokenBucket{
			tokens:     float64(rl.burst - 1),
			lastRefill: now,
		}
		return true
	}

	elapsed := now.Sub(bucket.lastRefill).Milliseconds()
	refill := int64(elapsed) * int64(rl.rps) / 1000
	bucket.tokens += float64(refill)
	if bucket.tokens > float64(rl.burst) {
		bucket.tokens = float64(rl.burst)
	}
	bucket.lastRefill = now

	if bucket.tokens >= 1 {
		bucket.tokens--
		return true
	}

	return false
}

var globalLimiter *RateLimiter

func RateLimit(rps, burst int) fiber.Handler {
	if globalLimiter == nil {
		globalLimiter = NewRateLimiter(rps, burst)
	}

	return func(c fiber.Ctx) error {
		key := c.IP()
		if !globalLimiter.Allow(key) {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "rate limit exceeded",
			})
		}
		return c.Next()
	}
}
