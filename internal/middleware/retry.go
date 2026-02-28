package middleware

import (
	"time"

	"github.com/gofiber/fiber/v3"
)

type RetryConfig struct {
	Attempts   int
	Backoff    time.Duration
	MaxBackoff time.Duration
}

func Retry(cfg RetryConfig) fiber.Handler {
	return func(c fiber.Ctx) error {
		if cfg.Attempts > 0 {
			c.Locals("retry_attempts", cfg.Attempts)
			c.Locals("retry_backoff", cfg.Backoff)
			c.Locals("retry_max_backoff", cfg.MaxBackoff)
		}

		return c.Next()
	}
}
