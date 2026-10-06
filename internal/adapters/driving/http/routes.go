package http

import "github.com/gofiber/fiber/v2"

func MapRoutes(app fiber.Router, h Handlers, authMiddleware fiber.Handler) {
	app.Post("/mortgage-profiles", authMiddleware, h.CreateCalculation())
	app.Get("/mortgage-profiles/:id", authMiddleware, h.GetCalculation())
}
