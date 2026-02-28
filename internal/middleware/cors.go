package middleware

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
)

type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	AllowCredentials bool
	ExposeHeaders    []string
	MaxAge           int
}

func CORS(config CORSConfig) fiber.Handler {
	defaultMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}
	defaultHeaders := []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"}

	if len(config.AllowMethods) == 0 {
		config.AllowMethods = defaultMethods
	}
	if len(config.AllowHeaders) == 0 {
		config.AllowHeaders = defaultHeaders
	}

	return func(c fiber.Ctx) error {
		origin := c.Get("Origin")

		allowed := false
		for _, o := range config.AllowOrigins {
			if o == "*" || o == origin {
				allowed = true
				break
			}
		}

		if allowed {
			if len(config.AllowOrigins) == 1 && config.AllowOrigins[0] == "*" {
				c.Set("Access-Control-Allow-Origin", "*")
			} else {
				c.Set("Access-Control-Allow-Origin", origin)
			}
			c.Set("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ", "))
			c.Set("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ", "))
			if config.AllowCredentials {
				c.Set("Access-Control-Allow-Credentials", "true")
			} else {
				c.Set("Access-Control-Allow-Credentials", "false")
			}

			if len(config.ExposeHeaders) > 0 {
				c.Set("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ", "))
			}
			if config.MaxAge > 0 {
				c.Set("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))
			}
		}

		if c.Method() == "OPTIONS" {
			return c.SendStatus(fiber.StatusNoContent)
		}

		return c.Next()
	}
}
