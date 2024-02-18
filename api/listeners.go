package api

import (
	"context"
	"fmt"
	"github.com/goccy/go-json"
	"strings"
	"time"
)

type Listeners struct {
	queries Queries
}

type (
	TransactionEvent struct {
		Value       int32     `json:"valor"`
		Type        string    `json:"tipo"`
		Description string    `json:"descricao"`
		CreatedAt   time.Time `json:"realizada_em"`
	}

	NewTransactionEvent struct {
		CustomerID  int32            `json:"customer_id"`
		Transaction TransactionEvent `json:"transaction"`
	}
)

func NewListeners(queries Queries) Listeners {
	return Listeners{
		queries: queries,
	}
}

func (l Listeners) SetupListeners() {
	go l.listenNewTransactions()
}

func (l Listeners) listenNewTransactions() {
	pool := l.queries.pool

	conn, err := pool.Acquire(context.Background())
	if err != nil {
		panic("Error acquiring connection:" + err.Error())
	}
	defer conn.Release()

	_, err = conn.Exec(context.Background(), "listen transaction_added")
	if err != nil {
		panic("Error listening to chat channel:" + err.Error())
	}
	for {
		notification, err := conn.Conn().WaitForNotification(context.Background())
		if err != nil {
			panic("Error waiting for notification:" + err.Error())
		}
		l.onNewTransaction(notification.Payload)
	}
}

func (l Listeners) onNewTransaction(payload string) {
	reader := strings.NewReader(payload)
	var event NewTransactionEvent
	if err := json.NewDecoder(reader).Decode(&event); err != nil {
		fmt.Println("onNewTransaction failed", err)
		return
	}
	go func() {
		if err := l.queries.InsertTransaction(context.Background(), event); err != nil {
			fmt.Println("onNewTransaction InsertTransaction failed", err)
		}
	}()
}
