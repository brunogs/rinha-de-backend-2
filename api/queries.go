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
	callCredit         = "SELECT c($1, $2, $3)"
	callDebit          = "SELECT d($1, $2, $3, $4)"
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
	selectCarteira = `
		SELECT valor, now() as realizada_em, ultimas_transacoes FROM carteiras 
		WHERE cliente_id = $1
	`
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
	row := q.pool.QueryRow(ctx, callCredit, customer.ID, transaction.Value, transaction.Description)
	var balance int32
	if err := row.Scan(&balance); err != nil {
		return 0, err
	}
	return balance, nil
}

func (q Queries) Debit(ctx context.Context, customer *Customer, transaction Transaction) (int32, error) {
	row := q.pool.QueryRow(ctx, callDebit, customer.ID, customer.Limit*-1, transaction.Value, transaction.Description)
	var balance int32
	if err := row.Scan(&balance); err != nil {
		return 0, ErrNoLimit
	}
	return balance, nil
}

func (q Queries) Extract(ctx context.Context, customerID int32) ([]ExtractRow, error) {
	rows, err := q.pool.Query(ctx, selectExtract, customerID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[ExtractRow])
}

func (q Queries) ExtractV2(ctx context.Context, customerID int32) (*ExtractOutput, error) {
	eo := ExtractOutput{}
	row := q.pool.QueryRow(ctx, selectCarteira, customerID)
	if err := row.Scan(&eo.Balance.Total, &eo.Balance.Date, &eo.LastTransactions); err != nil {
		return nil, err
	}
	return &eo, nil
}
