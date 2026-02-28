package middleware

import (
	"context"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v3"

	"api-gateway/internal/adapter/ratelimit"
)

var (
	limiters   = make(map[string]*ratelimit.TokenBucket)
	limitersMu sync.RWMutex
)

type RateLimitConfig struct {
	RouteID     string
	RouteRPS    int
	RouteBurst  int
	RouteKeyBy  string
	GlobalRPS   int
	GlobalBurst int
	GlobalKeyBy string
}

func RateLimit(rps, burst int) fiber.Handler {
	return RateLimitWithConfig(RateLimitConfig{
		RouteID:    "legacy",
		RouteRPS:   rps,
		RouteBurst: burst,
		RouteKeyBy: "ip",
	})
}

func RateLimitWithConfig(cfg RateLimitConfig) fiber.Handler {
	return func(c fiber.Ctx) error {
		if cfg.GlobalRPS > 0 && cfg.GlobalBurst > 0 {
			globalLimiter := getLimiter("global", cfg.GlobalRPS, cfg.GlobalBurst)
			globalKey := buildRateLimitKey(c, cfg.GlobalKeyBy)
			if allowed, err := globalLimiter.Allow(context.Background(), globalKey); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "rate limit error",
				})
			} else if !allowed {
				return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
					"error":       "global rate limit exceeded",
					"retry_after": "1s",
				})
			}
		}

		if cfg.RouteRPS > 0 && cfg.RouteBurst > 0 {
			routeID := cfg.RouteID
			if routeID == "" {
				routeID = c.Route().Path
			}

			routeLimiter := getLimiter("route:"+routeID, cfg.RouteRPS, cfg.RouteBurst)
			routeKey := buildRateLimitKey(c, cfg.RouteKeyBy)
			allowed, err := routeLimiter.Allow(context.Background(), routeKey)
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
		}

		return c.Next()
	}
}

func getLimiter(id string, rps, burst int) *ratelimit.TokenBucket {
	limitersMu.RLock()
	limiter, exists := limiters[id]
	limitersMu.RUnlock()

	if exists {
		return limiter
	}

	limitersMu.Lock()
	defer limitersMu.Unlock()

	if limiter, exists = limiters[id]; exists {
		return limiter
	}

	limiter = ratelimit.NewTokenBucket(rps, burst)
	limiters[id] = limiter
	return limiter
}

func buildRateLimitKey(c fiber.Ctx, strategy string) string {
	mode := strings.ToLower(strings.TrimSpace(strategy))
	switch mode {
	case "global":
		return "global"
	case "user", "per-user":
		if userID := GetUserID(c); userID != "" {
			return "user:" + userID
		}
		return "ip:" + c.IP()
	case "ip", "":
		fallthrough
	default:
		return "ip:" + c.IP()
	}
}
