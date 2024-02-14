package api

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"net/http"
	"strconv"
)

type GinHandler struct {
	queries   Queries
	customers map[int32]*Customer
}

func NewGinHandler(queries Queries) GinHandler {
	customers, err := queries.SelectAllCustomers(context.Background())
	if err != nil {
		panic("failed to load customers" + err.Error())
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
	r.GET("/clientes/:id/extrato", h.handleExtractV2)
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

	customer := h.customers[int32(id)]
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

func (h GinHandler) credit(c *gin.Context, customer *Customer, transaction Transaction) {
	balance, err := h.queries.Credit(c.Request.Context(), customer, transaction)
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
	balance, err := h.queries.Debit(c.Request.Context(), customer, transaction)
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

func (h GinHandler) handleExtractV2(c *gin.Context) {
	idStr := c.Param(fieldID)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.Status(http.StatusUnprocessableEntity)
		return
	}

	customer, ok := h.customers[int32(id)]
	if !ok {
		c.Status(http.StatusNotFound)
		return
	}
	extract, err := h.queries.Extract(c.Request.Context(), int32(id))
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	extract.Balance.Limit = customer.Limit
	c.JSON(http.StatusOK, *extract)
}
