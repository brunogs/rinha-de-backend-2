package api

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"net/http"
	"strconv"
	"sync"
)

type GinHandler struct {
	queries        Queries
	customers      map[int32]*Customer
	extractRowPool *sync.Pool
	balancePool    *sync.Pool
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

	extractPool := sync.Pool{
		New: func() any {
			return ExtractRow{}
		},
	}
	balancePool := sync.Pool{
		New: func() any {
			return Balance{}
		},
	}
	return GinHandler{
		queries:        queries,
		customers:      customersByID,
		extractRowPool: &extractPool,
		balancePool:    &balancePool,
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
	b := h.balancePool.Get().(Balance)
	defer h.balancePool.Put(b)
	b.Limit = customer.Limit
	b.BalanceValue = balance
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
	b := h.balancePool.Get().(Balance)
	defer h.balancePool.Put(b)
	b.Limit = customer.Limit
	b.BalanceValue = balance
	c.JSON(http.StatusOK, b)
}

func (h GinHandler) handleExtract(c *gin.Context) {
	idStr := c.Param(fieldID)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.Status(http.StatusUnprocessableEntity)
		return
	}

	extractRows, err := h.queries.Extract(c.Request.Context(), int32(id), h.extractRowPool)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	defer h.releaseRows(extractRows)

	customer := h.customers[int32(id)]
	if len(extractRows) > 0 {
		balance := extractRows[0]
		transactions := extractRows[1:]
		response := map[string]any{
			balanceLabel:         ExtractBalance{balance.Value, balance.CreatedAt, customer.Limit},
			lastTransactionLabel: transactions,
		}
		c.JSON(http.StatusOK, response)
		return
	}
	c.Status(http.StatusNotFound)
}

func (h GinHandler) releaseRows(rows []ExtractRow) {
	for _, row := range rows {
		h.extractRowPool.Put(row)
	}
}
