package api

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Queries struct {
	pool *pgxpool.Pool
}

func NewQueries(pool *pgxpool.Pool) Queries {
	return Queries{
		pool: pool,
	}
}

const (
	selectAllCustomers = "SELECT id, nome, limite FROM clientes;"
	callCredit         = "SELECT credit($1, $2, $3, $4)"
	callDebit          = "SELECT debit($1, $2, $3, $4, $5)"
	selectExtract      = `
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
	`

	selectBalance = "SELECT valor FROM saldos WHERE cliente_id = $1"
)

func (q Queries) SelectAllCustomers(ctx context.Context) ([]Customer, error) {
	rows, err := q.pool.Query(ctx, selectAllCustomers)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[Customer])
}

var ErrNoLimit = errors.New("Sem limite")

func (q Queries) Credit(ctx context.Context, customer *Customer, transaction Transaction) (int32, error) {
	_, err := q.pool.Exec(ctx, callCredit, customer.ID, transaction.Value, transaction.Type, transaction.Description)
	if err != nil {
		return 0, err
	}
	return q.balanceByCustomer(ctx, customer.ID)
}

func (q Queries) Debit(ctx context.Context, customer *Customer, transaction Transaction) (int32, error) {
	_, err := q.pool.Exec(ctx, callDebit, customer.ID, customer.Limit*-1, transaction.Value, transaction.Type, transaction.Description)
	if err != nil {
		return 0, ErrNoLimit
	}
	return q.balanceByCustomer(ctx, customer.ID)
}

func (q Queries) balanceByCustomer(ctx context.Context, customerID int32) (int32, error) {
	row := q.pool.QueryRow(ctx, selectBalance, customerID)
	var balance int32
	if err := row.Scan(&balance); err != nil {
		return 0, err
	}
	return balance, nil
}

func (q Queries) Extract(ctx context.Context, customerID int32) ([]ExtractRow, error) {
	rows, err := q.pool.Query(ctx, selectExtract, customerID)
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
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return extractRows, nil
}
