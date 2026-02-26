package middleware

import (
	"time"

	"github.com/gofiber/fiber/v3"
)

func Timeout(timeout time.Duration) fiber.Handler {
	return func(c fiber.Ctx) error {
		done := make(chan struct{})
		var err error

		go func() {
			err = c.Next()
			close(done)
		}()

		select {
		case <-done:
			return err
		case <-time.After(timeout):
			return c.Status(fiber.StatusGatewayTimeout).JSON(fiber.Map{
				"error": "request timeout",
			})
		}
	}
}
