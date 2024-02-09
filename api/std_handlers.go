package api

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type StdHandler struct {
	queries   Queries
	customers map[int32]*Customer
}

func NewStdHandler(queries Queries) StdHandler {
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
	return StdHandler{
		queries:   queries,
		customers: customersByID,
	}
}

func (h StdHandler) SetupEndpoints() {
	http.HandleFunc("/", h.handleRoot)
	http.HandleFunc("/clientes/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "transacoes") {
			h.handleTransaction(w, r)
			return
		} else if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "extrato") {
			h.handleExtract(w, r)
			return
		} else {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	})
}

func (h StdHandler) handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func (h StdHandler) handleTransaction(w http.ResponseWriter, r *http.Request) {
	var transaction Transaction
	err := json.NewDecoder(r.Body).Decode(&transaction)
	if err != nil || !transaction.isValid() {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	pathPieces := strings.Split(r.URL.Path, "/")
	idStr := pathPieces[2]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	customer := h.getCustomer(int32(id))
	if customer == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	switch transaction.Type {
	case credit:
		h.credit(w, customer, transaction)
	case debit:
		h.debit(w, customer, transaction)
	default:
		w.WriteHeader(http.StatusUnprocessableEntity)
	}
}

func (h StdHandler) getCustomer(customerID int32) *Customer {
	return h.customers[customerID]
}

func (h StdHandler) credit(w http.ResponseWriter, customer *Customer, transaction Transaction) {
	balance, err := h.queries.Credit(context.Background(), customer, transaction)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	b := Balance{
		Limit:        customer.Limit,
		BalanceValue: balance,
	}
	json.NewEncoder(w).Encode(b)
}

func (h StdHandler) debit(w http.ResponseWriter, customer *Customer, transaction Transaction) {
	balance, err := h.queries.Debit(context.Background(), customer, transaction)
	if err != nil {
		if errors.Is(err, ErrNoLimit) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	b := Balance{
		Limit:        customer.Limit,
		BalanceValue: balance,
	}
	json.NewEncoder(w).Encode(b)
}

func (h StdHandler) handleExtract(w http.ResponseWriter, r *http.Request) {
	pathPieces := strings.Split(r.URL.Path, "/")
	idStr := pathPieces[2]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	extractRows, err := h.queries.Extract(context.Background(), int32(id))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
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
		json.NewEncoder(w).Encode(response)
		return
	}
	w.WriteHeader(http.StatusNotFound)
}
