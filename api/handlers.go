package api

import (
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	"log"
	"time"
)

type (
	Handler struct {
		queries   Queries
		app       *fiber.App
		customers map[int32]*Customer
	}

	Transaction struct {
		Value       int32  `json:"valor"`
		Type        string `json:"tipo"`
		Description string `json:"descricao"`
	}

	Customer struct {
		ID    int32  `json:"id"`
		Name  string `json:"nome"`
		Limit int32  `json:"limite"`
	}

	Balance struct {
		Limit        int32 `json:"limite"`
		BalanceValue int32 `json:"saldo"`
	}

	ExtractRow struct {
		CreatedAt time.Time `json:"realizado_em"`
		Transaction
	}
)

const (
	credit = "c"
	debit  = "d"
)

func (t Transaction) isValid() bool {
	if t.Type != "c" && t.Type != "d" {
		return false
	}
	if len(t.Description) < 1 || len(t.Description) > 10 {
		return false
	}
	if t.Value == 0 {
		return false
	}
	return true
}

func NewHandler(queries Queries, app *fiber.App) Handler {
	attemps := 4
	var customers []Customer
	var err error
	for i := 0; i < attemps; i++ {
		customers, err = queries.SelectAllCustomers(context.Background())
		if err != nil {
			log.Println(err, "retrying")
			time.Sleep(500 * time.Millisecond)
		}
		if customers != nil {
			break
		}
	}
	if len(customers) == 0 {
		log.Fatal("falhou ao carregar clientes", err)
	}

	customersByID := make(map[int32]*Customer)
	for _, c := range customers {
		customer := c
		customersByID[c.ID] = &customer
	}
	return Handler{
		queries:   queries,
		app:       app,
		customers: customersByID,
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
	var transaction Transaction
	if err := c.BodyParser(&transaction); err != nil {
		return c.Status(422).SendString("body invalido")
	}
	if !transaction.isValid() {
		return c.Status(422).SendString("valores do body invalido")
	}
	customerID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(422).SendString("id invalido")
	}
	customer := h.getCustomer(int32(customerID))
	if customer == nil {
		return c.SendStatus(404)
	}
	switch transaction.Type {
	case credit:
		return h.credit(c, customer, transaction)
	case debit:
		return h.debit(c, customer, transaction)
	default:
		return c.SendStatus(422)
	}
}

func (h Handler) getCustomer(customerID int32) *Customer {
	return h.customers[customerID]
}

func (h Handler) credit(c *fiber.Ctx, customer *Customer, transaction Transaction) error {
	balance, err := h.queries.Credit(c.Context(), customer, transaction)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	b := Balance{
		Limit:        customer.Limit,
		BalanceValue: balance,
	}
	return c.JSON(b)
}

func (h Handler) debit(c *fiber.Ctx, customer *Customer, transaction Transaction) error {
	balance, err := h.queries.Debit(c.Context(), customer, transaction)
	if err != nil {
		if errors.Is(err, ErrNoLimit) {
			return c.SendStatus(422)
		}
		return c.Status(500).SendString(err.Error())
	}
	b := Balance{
		Limit:        customer.Limit,
		BalanceValue: balance,
	}
	return c.JSON(b)
}

func (h Handler) handleExtract(c *fiber.Ctx) error {
	customerID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(422).SendString("id invalido")
	}
	extractRows, err := h.queries.Extract(c.Context(), int32(customerID))
	if err != nil {
		return err
	}
	customer := h.getCustomer(int32(customerID))
	if len(extractRows) > 0 {
		balance := extractRows[0]
		transactions := extractRows[1:]
		response := map[string]any{
			"saldo": map[string]any{
				"total":        balance.Value,
				"data_extrato": balance.CreatedAt,
				"limite":       customer.Limit,
			},
			"ultimas_transacoes": transactions,
		}
		return c.JSON(response)
	}
	return c.SendStatus(404)
}
