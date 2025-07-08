package service_test

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/nehciyy/intrapay/internal/service"
)

func newMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db, mock
}

func TestCreateAccount_Success(t *testing.T) {
	db, mock := newMockDB(t)
	svc := service.NewService()

	accountID := int64(1)
	initialBalance := 100.0

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO accounts(account_id, balance) VALUES($1, $2)`)).
		WithArgs(accountID, initialBalance).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := svc.CreateAccount(db, accountID, initialBalance)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestCreateAccount_DuplicateKeyError(t *testing.T) {
	db, mock := newMockDB(t)
	svc := service.NewService()

	accountID := int64(1)
	initialBalance := 100.0

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO accounts(account_id, balance) VALUES($1, $2)`)).
		WithArgs(accountID, initialBalance).
		WillReturnError(fmt.Errorf("duplicate key value violates unique constraint"))

	err := svc.CreateAccount(db, accountID, initialBalance)
	if err == nil {
		t.Fatal("expected error on duplicate key, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetAccount_Success(t *testing.T) {
	db, mock := newMockDB(t)
	svc := service.NewService()

	accountID := int64(1)
	expectedBalance := 250.5

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT balance FROM accounts WHERE account_id = $1`)).
		WithArgs(accountID).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(expectedBalance))

	balance, err := svc.GetAccount(db, accountID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if balance != expectedBalance {
		t.Errorf("expected balance %v, got %v", expectedBalance, balance)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetAccount_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	svc := service.NewService()

	accountID := int64(1)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT balance FROM accounts WHERE account_id = $1`)).
		WithArgs(accountID).
		WillReturnError(sql.ErrNoRows)

	_, err := svc.GetAccount(db, accountID)
	if err == nil {
		t.Fatal("expected error for missing account, got nil")
	}
	if !errors.Is(err, sql.ErrNoRows) && err.Error() != fmt.Sprintf("account with ID %d not found", accountID) {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestCreateTransaction_Success(t *testing.T) {
	db, mock := newMockDB(t)
	svc := service.NewService()

	sourceID := int64(1)
	destID := int64(2)
	amount := 100.0
	transactionID := int64(1234)

	mock.ExpectBegin()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT balance FROM accounts WHERE account_id = $1 FOR UPDATE`)).
		WithArgs(sourceID).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(200.0))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM accounts WHERE account_id = $1)`)).
		WithArgs(destID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE accounts SET balance = balance + $1 WHERE account_id = $2`)).
		WithArgs(-amount, sourceID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE accounts SET balance = balance + $1 WHERE account_id = $2`)).
		WithArgs(amount, destID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO transactions (source_account_id, destination_account_id, amount)
		VALUES ($1, $2, $3) RETURNING id
	`)).
		WithArgs(sourceID, destID, amount).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(transactionID))

	mock.ExpectCommit()

	id, err := svc.CreateTransaction(db, sourceID, destID, amount)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedID := fmt.Sprintf("%d", transactionID)
	if id != expectedID {
		t.Errorf("expected transaction id %s, got %v", expectedID, id)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestCreateTransaction_InsufficientBalance(t *testing.T) {
	db, mock := newMockDB(t)
	svc := service.NewService()

	sourceID := int64(1)
	destID := int64(2)
	amount := 100.0

	mock.ExpectBegin()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT balance FROM accounts WHERE account_id = $1 FOR UPDATE`)).
		WithArgs(sourceID).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(50.0))

	mock.ExpectRollback()

	_, err := svc.CreateTransaction(db, sourceID, destID, amount)
	if err == nil {
		t.Fatal("expected error for insufficient balance, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestCreateTransaction_DestinationAccountNotFound(t *testing.T) {
	db, mock := newMockDB(t)
	svc := service.NewService()

	sourceID := int64(1)
	destID := int64(2)
	amount := 100.0

	mock.ExpectBegin()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT balance FROM accounts WHERE account_id = $1 FOR UPDATE`)).
		WithArgs(sourceID).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(200.0))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM accounts WHERE account_id = $1)`)).
		WithArgs(destID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	mock.ExpectRollback()

	_, err := svc.CreateTransaction(db, sourceID, destID, amount)
	if err == nil {
		t.Fatal("expected error for missing destination account, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}