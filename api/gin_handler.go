package api

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"log"
	"net/http"
	"strconv"
	"time"
)

type GinHandler struct {
	queries   Queries
	customers map[int32]*Customer
}

func NewGinHandler(queries Queries) GinHandler {
	attempts := 4
	var customers []Customer
	var err error
	for i := 0; i < attempts; i++ {
		customers, err = queries.SelectAllCustomers(context.Background())
		if err != nil {
			log.Println(err, "retrying")
			time.Sleep(500 * time.Millisecond)
		}
		if len(customers) != 0 {
			break
		}
	}
	if len(customers) == 0 {
		log.Fatal("failed to load customers", err)
	}

	customersByID := make(map[int32]*Customer)
	for _, c := range customers {
		customer := c
		customersByID[c.ID] = &customer
	}
	return GinHandler{
		queries:   queries,
		customers: customersByID,
	}
}

func (h GinHandler) SetupEndpoints(r *gin.Engine) {
	r.GET("/", h.handleRoot)
	r.POST("/clientes/:id/transacoes", h.handleTransaction)
	r.GET("/clientes/:id/extrato", h.handleExtract)
}

func (h GinHandler) handleRoot(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}

func (h GinHandler) handleTransaction(c *gin.Context) {
	var transaction Transaction
	if err := c.ShouldBindWith(&transaction, binding.JSON); err != nil {
		c.AbortWithError(http.StatusUnprocessableEntity, err).SetType(gin.ErrorTypeBind)
		return
	}
	if !transaction.isValid() {
		c.Status(http.StatusUnprocessableEntity)
		return
	}

	idStr := c.Param(fieldID)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.Status(http.StatusUnprocessableEntity)
		return
	}

	customer := h.getCustomer(int32(id))
	if customer == nil {
		c.Status(http.StatusNotFound)
		return
	}

	switch transaction.Type {
	case credit:
		h.credit(c, customer, transaction)
	case debit:
		h.debit(c, customer, transaction)
	default:
		c.Status(http.StatusUnprocessableEntity)
	}
}

func (h GinHandler) getCustomer(customerID int32) *Customer {
	return h.customers[customerID]
}

func (h GinHandler) credit(c *gin.Context, customer *Customer, transaction Transaction) {
	balance, err := h.queries.Credit(context.Background(), customer, transaction)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	b := Balance{
		Limit:        customer.Limit,
		BalanceValue: balance,
	}
	c.JSON(http.StatusOK, b)
}

func (h GinHandler) debit(c *gin.Context, customer *Customer, transaction Transaction) {
	balance, err := h.queries.Debit(context.Background(), customer, transaction)
	if err != nil {
		if errors.Is(err, ErrNoLimit) {
			c.Status(http.StatusUnprocessableEntity)
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}
	b := Balance{
		Limit:        customer.Limit,
		BalanceValue: balance,
	}
	c.JSON(http.StatusOK, b)
}

func (h GinHandler) handleExtract(c *gin.Context) {
	idStr := c.Param(fieldID)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.Status(http.StatusUnprocessableEntity)
		return
	}

	extractRows, err := h.queries.Extract(context.Background(), int32(id))
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	customer := h.getCustomer(int32(id))
	if len(extractRows) > 0 {
		balance := extractRows[0]
		transactions := extractRows[1:]
		response := map[string]interface{}{
			balanceLabel:         ExtractBalance{balance.Value, balance.CreatedAt, customer.Limit},
			lastTransactionLabel: transactions,
		}
		c.JSON(http.StatusOK, response)
		return
	}
	c.Status(http.StatusNotFound)
}
