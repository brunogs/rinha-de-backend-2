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

	debitUpdate = `
		UPDATE carteiras
		SET valor = valor - $1,
			ultimas_transacoes = (
				SELECT json_build_object(
				   'valor', $1,
				   'tipo', 'd',
				   'descricao', $2::text,
				   'realizada_em', now()
			   ) || ultimas_transacoes)[:10]
		WHERE cliente_id = $4 AND valor - $1 > $3
		RETURNING valor;
	`

	creditUpdate = `
		UPDATE carteiras
		SET valor = valor + $1,
			ultimas_transacoes = (
			SELECT json_build_object(
			 'valor', $1,
			 'tipo', 'c',
			 'descricao', $2::text,
			 'realizada_em', now()
			) || ultimas_transacoes)[:10]
		WHERE cliente_id = $3
		RETURNING valor;
	`

	selectCarteira = `
		SELECT valor, now() as realizada_em, ultimas_transacoes 
		FROM carteiras
		WHERE cliente_id = $1
	`

	insertTransaction = `
		INSERT INTO transacoes (cliente_id, valor, tipo, descricao, realizada_em)
		VALUES ($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING;
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

func (q Queries) Credit(ctx context.Context, customer *Customer, transaction *Transaction) (int32, error) {
	row := q.pool.QueryRow(ctx, creditUpdate, transaction.Value, transaction.Description, customer.ID)
	var balance int32
	if err := row.Scan(&balance); err != nil {
		return 0, err
	}
	return balance, nil
}

func (q Queries) Debit(ctx context.Context, customer *Customer, transaction *Transaction) (int32, error) {
	row := q.pool.QueryRow(ctx, debitUpdate, transaction.Value, transaction.Description, customer.Limit*-1, customer.ID)
	var balance int32
	if err := row.Scan(&balance); err != nil {
		return 0, ErrNoLimit
	}
	return balance, nil
}

func (q Queries) Extract(ctx context.Context, customerID int32) (*ExtractOutput, error) {
	eo := ExtractOutput{}
	row := q.pool.QueryRow(ctx, selectCarteira, customerID)
	if err := row.Scan(&eo.Balance.Total, &eo.Balance.Date, &eo.LastTransactions); err != nil {
		return nil, err
	}
	return &eo, nil
}

func (q Queries) InsertTransaction(ctx context.Context, event NewTransactionEvent) error {
	t := event.Transaction
	_, err := q.pool.Exec(ctx, insertTransaction, event.CustomerID, t.Value, t.Type, t.Description, t.CreatedAt)
	return err
}
