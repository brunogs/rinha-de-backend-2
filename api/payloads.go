package api

import "time"

const (
	credit = "c"
	debit  = "d"

	fieldID = "id"
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

	ExtractOutput struct {
		Balance          ExtractBalance   `json:"saldo"`
		LastTransactions []map[string]any `json:"ultimas_transacoes"`
	}

	ReportBalanceRow struct {
		Customer string `json:"cliente"`
		Limit    int32  `json:"limite"`
		Sum      int32  `json:"soma_transacoes"`
		Balance  int32  `json:"saldo"`
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
