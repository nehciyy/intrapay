package service

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

type DefaultService struct{}
const maxRetries = 3


func (s *DefaultService) CreateAccount(db *sql.DB, accountID int64, initialBalance float64) error {
	query := `INSERT INTO accounts(account_id, balance) VALUES($1, $2)`
	_, err := db.Exec(query, accountID, initialBalance)
	return err
}

func (s *DefaultService) GetAccount(db *sql.DB, accountID int64) (float64, error) {
	var balance float64

	query := `SELECT balance FROM accounts WHERE account_id = $1`
	err := db.QueryRow(query, accountID).Scan(&balance)

	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("account with ID %d not found", accountID)
	}

	return balance, err
}

func (s *DefaultService) CreateTransaction(db *sql.DB, sourceID int64, destID int64, amount float64) (string, error) {
	var transactionID string

	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Begin a new transaction
		tx, err := db.Begin()
		if err != nil {
			return "", fmt.Errorf("failed to begin transaction: %w", err)
		}

		// Define a rollback function to handle transaction rollback
		rollback := func(cause string) {
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Printf("rollback failed: %v", rbErr)
			}
			log.Printf("transaction rolled back due to: %s", cause)
		}

		// Check if source account exists and has sufficient balance
		sourceBalance, err := getBalanceTx(tx, sourceID)
		if err != nil {
			rollback(fmt.Sprintf("error retrieving source account: %v", err))
			return "", err
		}
		if sourceBalance < amount {
			rollback(fmt.Sprintf("insufficient balance in account %d", sourceID))
			return "", fmt.Errorf("insufficient balance in account %d", sourceID)
		}

		// Check if destination account exists
		destExists, err := accountExistsTx(tx, destID)
		if err != nil {
			rollback("error checking destination account: " + err.Error())
			return "", err
		}
		if !destExists {
			rollback(fmt.Sprintf("destination account %d not found", destID))
			return "", fmt.Errorf("destination account %d not found", destID)
		}

		// Deduct amount from source account balance
		err = updateBalanceTx(tx, sourceID, -amount)
		if err != nil {
			rollback("error updating source balance: " + err.Error())
			return "", err
		}

		// Add amount to destination account balance
		err = updateBalanceTx(tx, destID, amount)
		if err != nil {
			rollback("error updating destination balance: " + err.Error())
			return "", err
		}

		// Insert transaction record and get the new transaction ID
		transactionID, err = insertTransactionLogTx(tx, sourceID, destID, amount)
		if err != nil {
			rollback("error inserting transaction record: " + err.Error())
			return "", err
		}

		// Commit the transaction
		if err := tx.Commit(); err != nil {
			// If the error is a serialization failure (deadlock), we can retry until maxRetries
			if isSerializationFailure(err) && attempt < maxRetries {
				log.Printf("serialization failure, retrying attempt %d...", attempt)
				time.Sleep(100 * time.Millisecond) // Sleep before retrying to avoid immediate contention
				continue
			}
			// For other errors, rollback and return the error
			rollback("commit failed: " + err.Error())
			return "", err
		}

		break // Exit the retry loop if commit was successful
	}

	// If after retries no transaction ID was genereated, report failure
	if transactionID == "" {
		return "", errors.New("transaction failed after max retries")
	}

	return transactionID, nil
}

func accountExistsTx(tx *sql.Tx, accountID int64) (bool, error) {
	var exists bool
	err := tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM accounts WHERE account_id = $1)`, accountID).Scan(&exists)
	return exists, err
}

func getBalanceTx(tx *sql.Tx, accountID int64) (float64, error) {
	var balance float64
	err := tx.QueryRow(`SELECT balance FROM accounts WHERE account_id = $1 FOR UPDATE`, accountID).Scan(&balance) // Use FOR UPDATE to lock the row, to prevent race conditions from simultaneous transactions
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("account with ID %d not found", accountID)
	}
	return balance, err
}

func updateBalanceTx(tx *sql.Tx, accountID int64, delta float64) error {
	query := `UPDATE accounts SET balance = balance + $1 WHERE account_id = $2`
	_, err := tx.Exec(query, delta, accountID)
	return err
}

func insertTransactionLogTx(tx *sql.Tx, sourceID, destID int64, amount float64) (string, error) {
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

func isSerializationFailure(err error) bool {
	return err != nil && strings.Contains(err.Error(), "SQLSTATE 40001")
}