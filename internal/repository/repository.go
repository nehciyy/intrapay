package repository

import "database/sql"

// AccountRepository defines the interface for account-related database operations.
type AccountRepository interface {
	CreateAccount(accountID int64, initialBalance float64) error
	GetAccountBalance(accountID int64) (float64, error)
	AccountExists(accountID int64) (bool, error) // Added for transaction logic
}

// TransactionRepository defines the interface for transaction-related database operations.
type TransactionRepository interface {
	GetAccountBalanceTx(tx *sql.Tx, accountID int64) (float64, error)
	AccountExistsTx(tx *sql.Tx, accountID int64) (bool, error)
	UpdateBalanceTx(tx *sql.Tx, accountID int64, delta float64) error
	InsertTransactionLogTx(tx *sql.Tx, sourceID, destID int64, amount float64) (string, error)
}