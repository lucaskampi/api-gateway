package middleware

import (
	"time"

	"github.com/gofiber/fiber/v3"
)

func Timeout(timeout time.Duration) fiber.Handler {
	return func(c fiber.Ctx) error {
		if timeout > 0 {
			c.Locals("request_timeout", timeout)
		}
		return c.Next()
	}
}
