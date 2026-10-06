package http

import (
	"errors"

	"github.com/armbdevelop/testgomg/internal/core/domain"
	"github.com/gofiber/fiber/v2"
)

const errKey = "error"

func ErrorHandler(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domain.ErrValidation):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{errKey: err.Error()})
	case errors.Is(err, domain.ErrNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{errKey: err.Error()})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{errKey: "внутренняя ошибка сервера"})
	}
}
