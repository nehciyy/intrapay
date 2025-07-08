package service

import (
	"database/sql"
	"fmt"
)

func CreateAccount(db *sql.DB, accountID int64, initialBalance float64) error {
	query := `INSERT INTO accounts(account_id, balance) VALUES($1, $2)`
	_, err := db.Exec(query, accountID, initialBalance)
	return err
}

func GetAccount(db *sql.DB, accountID int64) (float64, error) {
	var balance float64

	query := `SELECT balance FROM accounts WHERE account_id = $1`
	err := db.QueryRow(query, accountID).Scan(&balance)

	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("account with ID %d not found", accountID)
	}

	return balance, err
}
