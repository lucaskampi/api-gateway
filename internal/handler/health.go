package handler

import (
	"github.com/gofiber/fiber/v3"
)

func Health() fiber.Handler {
	return func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	}
}

func Ready() fiber.Handler {
	return func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ready",
		})
	}
}
