package api

import "time"

const (
	credit = "c"
	debit  = "d"

	fieldID = "id"

	balanceLabel         = "saldo"
	lastTransactionLabel = "ultimas_transacoes"
)

type (
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

	ExtractBalance struct {
		Total int32     `json:"total"`
		Date  time.Time `json:"data_extrato"`
		Limit int32     `json:"limite"`
	}
	ExtractRow struct {
		CreatedAt time.Time `json:"realizado_em"`
		Transaction
	}
)

func (t Transaction) isValid() bool {
	if len(t.Description) < 1 || len(t.Description) > 10 {
		return false
	}
	if t.Value == 0 {
		return false
	}
	return true
}
