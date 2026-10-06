package http

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

const userIDKey = "userID"

func NewAuthMiddleware(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get(fiber.HeaderAuthorization)

		tokenString, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{errKey: "требуется токен авторизации"})
		}

		token, err := jwt.Parse(tokenString, func(_ *jwt.Token) (any, error) {
			return []byte(secret), nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{errKey: "некорректный токен"})
		}

		sub, err := token.Claims.GetSubject()
		if err != nil || sub == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{errKey: "в токене нет sub"})
		}

		c.Locals(userIDKey, sub)

		return c.Next()
	}
}

func userIDFrom(c *fiber.Ctx) (string, error) {
	userID, ok := c.Locals(userIDKey).(string)
	if !ok || userID == "" {
		return "", errors.New("userID not found in context")
	}

	return userID, nil
}
