package handler

import (
	"github.com/gofiber/fiber/v3"
)

type ReloadHandler func() interface{}

func Reload(reload ReloadHandler) fiber.Handler {
	return func(c fiber.Ctx) error {
		result := reload()
		if result == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to reload config",
			})
		}
		return c.JSON(fiber.Map{
			"status": "config reloaded successfully",
		})
	}
}
