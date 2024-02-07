package api

import "github.com/gofiber/fiber/v2"

type Handler struct {
	queries Queries
	app     *fiber.App
}

func NewHandler(queries Queries, app *fiber.App) Handler {
	return Handler{
		queries: queries,
		app:     app,
	}
}

func (h Handler) SetupEndpoints() {
	h.app.Get("/", h.handleRoot)
	h.app.Post("/clientes/:id/transacoes", h.handleTransaction)
	h.app.Get("/clientes/:id/extrato", h.handleExtract)
}

func (h Handler) handleRoot(c *fiber.Ctx) error {
	return c.SendString("OK")
}

func (h Handler) handleTransaction(c *fiber.Ctx) error {
	return nil
}

func (h Handler) handleExtract(c *fiber.Ctx) error {
	return nil
}
