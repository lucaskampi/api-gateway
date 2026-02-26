package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

const UserIDCtxKey = "user_id"
const UserClaimsCtxKey = "user_claims"

type JWTConfig struct {
	Secret string
	Issuer string
}

func JWT(config JWTConfig) fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing authorization header",
			})
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid authorization header format",
			})
		}

		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.ErrUnauthorized
			}
			return []byte(config.Secret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid token",
			})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid token claims",
			})
		}

		if config.Issuer != "" {
			iss, ok := claims["iss"].(string)
			if !ok || iss != config.Issuer {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "invalid issuer",
				})
			}
		}

		if userID, ok := claims["sub"].(string); ok {
			c.Locals(UserIDCtxKey, userID)
		}
		c.Locals(UserClaimsCtxKey, claims)

		return c.Next()
	}
}

func GetUserID(c fiber.Ctx) string {
	if id, ok := c.Locals(UserIDCtxKey).(string); ok {
		return id
	}
	return ""
}

func GetUserClaims(c fiber.Ctx) map[string]interface{} {
	if claims, ok := c.Locals(UserClaimsCtxKey).(map[string]interface{}); ok {
		return claims
	}
	return nil
}
