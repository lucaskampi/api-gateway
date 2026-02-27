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
		upstream := c.Locals("upstream")
		if upstream == nil {
			return c.Next()
		}

		var lastErr error
		backoff := cfg.Backoff

		for attempt := 0; attempt <= cfg.Attempts; attempt++ {
			if attempt > 0 {
				time.Sleep(backoff)
				backoff = backoff * 2
				if backoff > cfg.MaxBackoff {
					backoff = cfg.MaxBackoff
				}
			}

			lastErr = c.Next()
			if lastErr == nil {
				return nil
			}

			status := c.Response().StatusCode()
			if status >= 200 && status < 500 {
				return nil
			}
		}

		if lastErr != nil {
			return lastErr
		}

		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "all retry attempts failed",
		})
	}
}
