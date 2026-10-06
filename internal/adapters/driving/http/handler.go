package http

import (
	"context"
	"strconv"
	"time"

	"github.com/armbdevelop/testgomg/internal/core/domain"
	"github.com/armbdevelop/testgomg/internal/core/ports"
	"github.com/gofiber/fiber/v2"
)

const requestTimeout = 10 * time.Second

type Handlers interface {
	CreateCalculation() fiber.Handler
	GetCalculation() fiber.Handler
}

type handlers struct {
	svc ports.Service
}

func NewHandlers(svc ports.Service) Handlers { return &handlers{svc: svc} }

func (h *handlers) CreateCalculation() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req domain.MortgageProfile
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{errKey: "некорректный JSON"})
		}

		userID, err := userIDFrom(c)
		if err != nil {
			return err
		}

		req.UserID = userID

		ctx, cancel := context.WithTimeout(c.UserContext(), requestTimeout)
		defer cancel()

		id, err := h.svc.CreateCalculation(ctx, req)
		if err != nil {
			return err
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": strconv.FormatInt(id, 10)})
	}
}

func (h *handlers) GetCalculation() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{errKey: "id должен быть числом"})
		}

		userID, err := userIDFrom(c)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(c.UserContext(), requestTimeout)
		defer cancel()

		calc, err := h.svc.GetCalculation(ctx, id, userID)
		if err != nil {
			return err
		}

		return c.JSON(newCalculationResponse(calc))
	}
}
