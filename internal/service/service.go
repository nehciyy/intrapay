package service

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/nehciyy/intrapay/internal/repository"
)

type DefaultService struct {
	accountRepo     repository.AccountRepository
	transactionRepo repository.TransactionRepository
	db              *sql.DB
}

func NewService(db *sql.DB, accountRepo repository.AccountRepository, transactionRepo repository.TransactionRepository) Service {
	return &DefaultService{
		db:              db,
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
	}
}

const maxRetries = 3

func (s *DefaultService) CreateAccount(accountID int64, initialBalance float64) error {
	return s.accountRepo.CreateAccount(accountID, initialBalance)
}

func (s *DefaultService) GetAccount(accountID int64) (float64, error) {
	return s.accountRepo.GetAccountBalance(accountID)
}

func (s *DefaultService) CreateTransaction(sourceID int64, destID int64, amount float64) (string, error) {
	var transactionID string

	for attempt := 1; attempt <= maxRetries; attempt++ {
		tx, err := s.db.Begin()
		if err != nil {
			return "", fmt.Errorf("failed to begin transaction: %w", err)
		}

		rolledBack := false
		rollback := func(cause string) {
			if rolledBack {
				return
			}
			if rbErr := tx.Rollback(); rbErr != nil && rbErr.Error() != "sql: transaction has already been committed or rolled back" {
				log.Printf("rollback failed: %v", rbErr)
			}
			log.Printf("transaction rolled back due to: %s", cause)
			rolledBack = true
		}

		sourceBalance, err := s.transactionRepo.GetAccountBalanceTx(tx, sourceID)
		if err != nil {
			rollback(fmt.Sprintf("error retrieving source account: %v", err))
			return "", err
		}
		if sourceBalance < amount {
			rollback(fmt.Sprintf("insufficient balance in account %d", sourceID))
			return "", fmt.Errorf("insufficient balance in account %d", sourceID)
		}

		destExists, err := s.transactionRepo.AccountExistsTx(tx, destID)
		if err != nil {
			rollback("error checking destination account: " + err.Error())
			return "", err
		}
		if !destExists {
			rollback(fmt.Sprintf("destination account %d not found", destID))
			return "", fmt.Errorf("destination account %d not found", destID)
		}

		if err := s.transactionRepo.UpdateBalanceTx(tx, sourceID, -amount); err != nil {
			rollback("error updating source balance: " + err.Error())
			return "", err
		}
		if err := s.transactionRepo.UpdateBalanceTx(tx, destID, amount); err != nil {
			rollback("error updating destination balance: " + err.Error())
			return "", err
		}

		transactionID, err = s.transactionRepo.InsertTransactionLogTx(tx, sourceID, destID, amount)
		if err != nil {
			rollback("error inserting transaction record: " + err.Error())
			return "", err
		}

		err = tx.Commit()
		if err != nil {
			if repository.IsSerializationFailure(err) {
				log.Printf("serialization failure, retrying attempt %d...", attempt)
				time.Sleep(100 * time.Millisecond)
				continue
			}
			rollback(fmt.Sprintf("commit failed: %v", err))
			return "", fmt.Errorf("commit failed: %v", err)
		}

		return transactionID, nil
	}

	return "", errors.New("transaction failed after max retries")
}