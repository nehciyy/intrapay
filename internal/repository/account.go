package repository

import (
	"database/sql"
	"fmt"
	"strings"
)

// PostgresAccountRepository is an implementation of AccountRepository for PostgreSQL.
type PostgresAccountRepository struct {
	db *sql.DB
}

type PostgresTransactionRepository struct {
	db *sql.DB
}

func NewPostgresTransactionRepository(db *sql.DB) *PostgresTransactionRepository {
	return &PostgresTransactionRepository{db: db}
}

// NewPostgresAccountRepository creates a new PostgresAccountRepository.
func NewPostgresAccountRepository(db *sql.DB) *PostgresAccountRepository {
	return &PostgresAccountRepository{db: db}
}

func (r *PostgresAccountRepository) CreateAccount(accountID int64, initialBalance float64) error {
	query := `INSERT INTO accounts(account_id, balance) VALUES($1, $2)`
	_, err := r.db.Exec(query, accountID, initialBalance)
	return err
}

func (r *PostgresAccountRepository) GetAccountBalance(accountID int64) (float64, error) {
	var balance float64
	query := `SELECT balance FROM accounts WHERE account_id = $1`
	err := r.db.QueryRow(query, accountID).Scan(&balance)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("account with ID %d not found", accountID)
	}
	return balance, err
}

func (r *PostgresAccountRepository) AccountExists(accountID int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM accounts WHERE account_id = $1)`, accountID).Scan(&exists)
	return exists, err
}


func (r *PostgresTransactionRepository) GetAccountBalanceTx(tx *sql.Tx, accountID int64) (float64, error) {
	var balance float64
	// Use FOR UPDATE to lock the row, to prevent race conditions from simultaneous transactions
	err := tx.QueryRow(`SELECT balance FROM accounts WHERE account_id = $1 FOR UPDATE`, accountID).Scan(&balance)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("account with ID %d not found", accountID)
	}
	return balance, err
}

func (r *PostgresTransactionRepository) AccountExistsTx(tx *sql.Tx, accountID int64) (bool, error) {
	var exists bool
	err := tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM accounts WHERE account_id = $1)`, accountID).Scan(&exists)
	return exists, err
}

func (r *PostgresTransactionRepository) UpdateBalanceTx(tx *sql.Tx, accountID int64, delta float64) error {
	query := `UPDATE accounts SET balance = balance + $1 WHERE account_id = $2`
	_, err := tx.Exec(query, delta, accountID)
	return err
}

func (r *PostgresTransactionRepository) InsertTransactionLogTx(tx *sql.Tx, sourceID, destID int64, amount float64) (string, error) {
	var id int64
	err := tx.QueryRow(`
		INSERT INTO transactions (source_account_id, destination_account_id, amount)
		VALUES ($1, $2, $3) RETURNING id
	`, sourceID, destID, amount).Scan(&id)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d", id), nil
}

// isSerializationFailure checks if the error is a PostgreSQL serialization failure (SQLSTATE 40001).
func IsSerializationFailure(err error) bool {
	return err != nil && strings.Contains(err.Error(), "SQLSTATE 40001")
}