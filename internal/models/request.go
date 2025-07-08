package models

type CreateAccountRequest struct {
	AccountID      int64   `json:"account_id"`
	InitialBalance float64 `json:"initial_balance"`
}

type TransactionRequest struct {
	SourceAccountID      int64   `json:"source_account_id"`
	DestinationAccountID int64   `json:"destination_account_id"`
	Amount               float64 `json:"amount"`
}