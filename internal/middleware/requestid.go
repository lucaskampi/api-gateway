package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

const RequestIDHeader = "X-Request-ID"
const RequestIDCtxKey = "request_id"

func RequestID() fiber.Handler {
	return func(c fiber.Ctx) error {
		reqID := c.Get(RequestIDHeader)
		if reqID == "" {
			reqID = uuid.New().String()
		}
		c.Set(RequestIDHeader, reqID)
		c.Locals(RequestIDCtxKey, reqID)
		return c.Next()
	}
}

func GetRequestID(c fiber.Ctx) string {
	if id, ok := c.Locals(RequestIDCtxKey).(string); ok {
		return id
	}
	return ""
}
