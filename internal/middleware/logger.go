package middleware

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
)

func Logger(logger zerolog.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		duration := time.Since(start)

		log := logger.With().
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", c.Response().StatusCode()).
			Dur("duration", duration).
			Str("request_id", GetRequestID(c)).
			Str("remote_ip", c.IP()).
			Logger()

		if err != nil {
			log.Error().Err(err).Msg("request failed")
		} else {
			log.Info().Msg("request completed")
		}

		return err
	}
}
