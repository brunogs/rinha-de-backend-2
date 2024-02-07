package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
)

type Queries struct {
	pool *pgxpool.Pool
}

func NewQueries(pool *pgxpool.Pool) Queries {
	return Queries{
		pool: pool,
	}
}

func (q Queries) SelectAllCustomers(ctx context.Context) ([]Customer, error) {
	rows, err := q.pool.Query(ctx, "select id, nome, limite from clientes;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var customers []Customer
	for rows.Next() {
		var c Customer
		err = rows.Scan(&c.ID, &c.Name, &c.Limit)
		if err != nil {
			return nil, err
		}
		customers = append(customers, c)
	}
	return customers, nil
}

var ErrNoLimit = errors.New("Sem limite")

func (q Queries) Credit(ctx context.Context, customer *Customer, transaction Transaction) (int32, error) {
	tx, err := q.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, `SELECT credit($1, $2, $3, $4)`, customer.ID, transaction.Value, transaction.Type, transaction.Description)
	var balance int32
	err = row.Scan(&balance)
	if err != nil {
		fmt.Println(err)
		return 0, ErrNoLimit
	}
	if err = tx.Commit(ctx); err != nil {
		return 0, err
	}
	return balance, nil
}

func (q Queries) Debit(ctx context.Context, customer *Customer, transaction Transaction) (int32, error) {
	tx, err := q.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, `SELECT debit($1, $2, $3, $4, $5)`, customer.ID, customer.Limit*-1, transaction.Value, transaction.Type, transaction.Description)
	var balance int32
	err = row.Scan(&balance)
	if err != nil {
		if strings.Contains(err.Error(), "Saldo insuficiente") {
			return 0, ErrNoLimit
		}
		return 0, err
	}
	if err = tx.Commit(ctx); err != nil {
		return 0, err
	}
	return balance, nil
}

func (q Queries) Extract(ctx context.Context, customerID int32) ([]ExtractRow, error) {
	rows, err := q.pool.Query(ctx, `
		 (SELECT 
		     valor, 
		     'saldo' AS tipo, 
		     'saldo' AS descricao, 
		     now() AS realizada_em
		 FROM saldos
		 WHERE cliente_id = $1)
		 UNION ALL
		 (SELECT 
			valor, 
			tipo, 
			descricao, 
			realizada_em
		 FROM transacoes
		 WHERE cliente_id = $1
		 ORDER BY id desc LIMIT 10)
	`, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var extractRows []ExtractRow
	for rows.Next() {
		var e ExtractRow
		err = rows.Scan(&e.Value, &e.Type, &e.Description, &e.CreatedAt)
		if err != nil {
			return nil, err
		}
		extractRows = append(extractRows, e)
	}
	return extractRows, nil
}
