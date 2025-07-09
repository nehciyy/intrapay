package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err, "sqlmock.New should not return an error")
	t.Cleanup(func() {
		db.Close() // Ensure the mock DB is closed after the test
	})
	return db, mock
}

// TestNewPostgresAccountRepository tests the constructor for the repository.
func TestNewPostgresAccountRepository(t *testing.T) {
	db, _ := setupMockDB(t)

	repo := NewPostgresAccountRepository(db)
	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}

// TestCreateAccount tests the CreateAccount method.
func TestPostgresAccountRepository_CreateAccount(t *testing.T) {
	db, mock := setupMockDB(t)

	repo := NewPostgresAccountRepository(db)

	tests := []struct {
		name          string
		accountID     int64
		initialBalance float64
		mockExpect    func()
		expectedError error
	}{
		{
			name:          "Successful creation",
			accountID:     1001,
			initialBalance: 500.00,
			mockExpect: func() {
				mock.ExpectExec("INSERT INTO accounts").
					WithArgs(int64(1001), 500.00).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedError: nil,
		},
		{
			name:          "Database error",
			accountID:     1002,
			initialBalance: 200.00,
			mockExpect: func() {
				mock.ExpectExec("INSERT INTO accounts").
					WithArgs(int64(1002), 200.00).
					WillReturnError(errors.New("db connection error"))
			},
			expectedError: errors.New("db connection error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockExpect()
			err := repo.CreateAccount(tt.accountID, tt.initialBalance)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet()) // Verify all expectations were met
		})
	}
}

// TestGetAccountBalance tests the GetAccountBalance method.
func TestPostgresAccountRepository_GetAccountBalance(t *testing.T) {
	db, mock := setupMockDB(t)

	repo := NewPostgresAccountRepository(db)

	tests := []struct {
		name          string
		accountID     int64
		mockExpect    func()
		expectedBalance float64
		expectedError error
	}{
		{
			name:          "Successful retrieval",
			accountID:     1001,
			mockExpect: func() {
				rows := sqlmock.NewRows([]string{"balance"}).AddRow(1000.50)
				mock.ExpectQuery("SELECT balance FROM accounts WHERE account_id = \\$1").
					WithArgs(int64(1001)).
					WillReturnRows(rows)
			},
			expectedBalance: 1000.50,
			expectedError: nil,
		},
		{
			name:          "Account not found",
			accountID:     1002,
			mockExpect: func() {
				mock.ExpectQuery("SELECT balance FROM accounts WHERE account_id = \\$1").
					WithArgs(int64(1002)).
					WillReturnError(sql.ErrNoRows)
			},
			expectedBalance: 0,
			expectedError: fmt.Errorf("account with ID %d not found", 1002),
		},
		{
			name:          "Database error",
			accountID:     1003,
			mockExpect: func() {
				mock.ExpectQuery("SELECT balance FROM accounts WHERE account_id = \\$1").
					WithArgs(int64(1003)).
					WillReturnError(errors.New("query failed"))
			},
			expectedBalance: 0,
			expectedError: errors.New("query failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockExpect()
			balance, err := repo.GetAccountBalance(tt.accountID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBalance, balance)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestAccountExists tests the AccountExists method.
func TestPostgresAccountRepository_AccountExists(t *testing.T) {
	db, mock := setupMockDB(t)

	repo := NewPostgresAccountRepository(db)

	tests := []struct {
		name          string
		accountID     int64
		mockExpect    func()
		expectedExists bool
		expectedError error
	}{
		{
			name:          "Account exists",
			accountID:     1001,
			mockExpect: func() {
				rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
				mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM accounts WHERE account_id = \\$1\\)").
					WithArgs(int64(1001)).
					WillReturnRows(rows)
			},
			expectedExists: true,
			expectedError: nil,
		},
		{
			name:          "Account does not exist",
			accountID:     1002,
			mockExpect: func() {
				rows := sqlmock.NewRows([]string{"exists"}).AddRow(false)
				mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM accounts WHERE account_id = \\$1\\)").
					WithArgs(int64(1002)).
					WillReturnRows(rows)
			},
			expectedExists: false,
			expectedError: nil,
		},
		{
			name:          "Database error",
			accountID:     1003,
			mockExpect: func() {
				mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM accounts WHERE account_id = \\$1\\)").
					WithArgs(int64(1003)).
					WillReturnError(errors.New("db error"))
			},
			expectedExists: false,
			expectedError: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockExpect()
			exists, err := repo.AccountExists(tt.accountID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedExists, exists)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestGetAccountBalanceTx tests the GetAccountBalanceTx method.
func TestPostgresAccountRepository_GetAccountBalanceTx(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewPostgresTransactionRepository(db)

	tests := []struct {
		name          string
		accountID     int64
		mockExpect    func(sqlmock.Sqlmock) *sql.Tx // Now returns *sql.Tx
		expectedBalance float64
		expectedError error
	}{
		{
			name:          "Successful retrieval in transaction",
			accountID:     1001,
			mockExpect: func(mock sqlmock.Sqlmock) *sql.Tx {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"balance"}).AddRow(500.00)
				mock.ExpectQuery("SELECT balance FROM accounts WHERE account_id = \\$1 FOR UPDATE").
					WithArgs(int64(1001)).
					WillReturnRows(rows)
				mock.ExpectRollback() // Expect rollback as we'll explicitly call it
				tx, _ := db.Begin()   // Start the actual mock transaction
				return tx
			},
			expectedBalance: 500.00,
			expectedError: nil,
		},
		{
			name:          "Account not found in transaction",
			accountID:     1002,
			mockExpect: func(mock sqlmock.Sqlmock) *sql.Tx {
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT balance FROM accounts WHERE account_id = \\$1 FOR UPDATE").
					WithArgs(int64(1002)).
					WillReturnError(sql.ErrNoRows)
				mock.ExpectRollback()
				tx, _ := db.Begin()
				return tx
			},
			expectedBalance: 0,
			expectedError: fmt.Errorf("account with ID %d not found", 1002),
		},
		{
			name:          "Database error in transaction",
			accountID:     1003,
			mockExpect: func(mock sqlmock.Sqlmock) *sql.Tx {
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT balance FROM accounts WHERE account_id = \\$1 FOR UPDATE").
					WithArgs(int64(1003)).
					WillReturnError(errors.New("tx query failed"))
				mock.ExpectRollback()
				tx, _ := db.Begin()
				return tx
			},
			expectedBalance: 0,
			expectedError: errors.New("tx query failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := tt.mockExpect(mock) // Get the mock transaction from mockExpect
			
			balance, err := repo.GetAccountBalanceTx(tx, tt.accountID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBalance, balance)
			}
			assert.NoError(t, tx.Rollback()) // Explicitly rollback the mock transaction
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestAccountExistsTx tests the AccountExistsTx method.
func TestPostgresAccountRepository_AccountExistsTx(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewPostgresTransactionRepository(db)

	tests := []struct {
		name          string
		accountID     int64
		mockExpect    func(sqlmock.Sqlmock, *sql.DB) *sql.Tx
		expectedExists bool
		expectedError error
	}{
		{
			name:          "Account exists in transaction",
			accountID:     1001,
			mockExpect: func(mock sqlmock.Sqlmock, db *sql.DB) *sql.Tx {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
				mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM accounts WHERE account_id = \\$1\\)").
					WithArgs(int64(1001)).
					WillReturnRows(rows)
				mock.ExpectRollback()
				tx, err := db.Begin()
				assert.NoError(t, err)
				return tx
			},
			expectedExists: true,
			expectedError: nil,
		},
		{
			name:          "Account does not exist in transaction",
			accountID:     1002,
			mockExpect: func(mock sqlmock.Sqlmock, db *sql.DB) *sql.Tx {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"exists"}).AddRow(false)
				mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM accounts WHERE account_id = \\$1\\)").
					WithArgs(int64(1002)).
					WillReturnRows(rows)
				mock.ExpectRollback()
				tx, err := db.Begin()
				assert.NoError(t, err)
				return tx
			},
			expectedExists: false,
			expectedError: nil,
		},
		{
			name:          "Database error in transaction",
			accountID:     1003,
			mockExpect: func(mock sqlmock.Sqlmock, db *sql.DB) *sql.Tx {
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM accounts WHERE account_id = \\$1\\)").
					WithArgs(int64(1003)).
					WillReturnError(errors.New("tx exists query failed"))
				mock.ExpectRollback()
				tx, err := db.Begin()
				assert.NoError(t, err)
				return tx
			},
			expectedExists: false,
			expectedError: errors.New("tx exists query failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := tt.mockExpect(mock, db)
			
			exists, err := repo.AccountExistsTx(tx, tt.accountID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedExists, exists)
			}
			assert.NoError(t, tx.Rollback())
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestUpdateBalanceTx tests the UpdateBalanceTx method.
func TestPostgresAccountRepository_UpdateBalanceTx(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewPostgresTransactionRepository(db)

	tests := []struct {
		name          string
		accountID     int64
		delta         float64
		mockExpect    func(sqlmock.Sqlmock, *sql.DB) *sql.Tx // Pass db here
		expectedError error
	}{
		{
			name:          "Successful balance update (add)",
			accountID:     1001,
			delta:         100.00,
			mockExpect: func(mock sqlmock.Sqlmock, db *sql.DB) *sql.Tx {
				mock.ExpectBegin() // Expect Begin for this transaction
				mock.ExpectExec("UPDATE accounts SET balance = balance \\+ \\$1 WHERE account_id = \\$2").
					WithArgs(100.00, int64(1001)).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectRollback() // Expect rollback as we'll explicitly call it
				tx, err := db.Begin() // Start the actual mock transaction
				assert.NoError(t, err) // Assert no error on mock Begin
				return tx
			},
			expectedError: nil,
		},
		{
			name:          "Successful balance update (deduct)",
			accountID:     1002,
			delta:         -50.00,
			mockExpect: func(mock sqlmock.Sqlmock, db *sql.DB) *sql.Tx {
				mock.ExpectBegin() // Expect Begin for this transaction
				mock.ExpectExec("UPDATE accounts SET balance = balance \\+ \\$1 WHERE account_id = \\$2").
					WithArgs(-50.00, int64(1002)).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectRollback()
				tx, err := db.Begin()
				assert.NoError(t, err)
				return tx
			},
			expectedError: nil,
		},
		{
			name:          "Database error during update",
			accountID:     1003,
			delta:         200.00,
			mockExpect: func(mock sqlmock.Sqlmock, db *sql.DB) *sql.Tx {
				mock.ExpectBegin() // Expect Begin for this transaction
				mock.ExpectExec("UPDATE accounts SET balance = balance \\+ \\$1 WHERE account_id = \\$2").
					WithArgs(200.00, int64(1003)).
					WillReturnError(errors.New("tx update failed"))
				mock.ExpectRollback()
				tx, err := db.Begin()
				assert.NoError(t, err)
				return tx
			},
			expectedError: errors.New("tx update failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := tt.mockExpect(mock, db) // Pass db to mockExpect
			
			err := repo.UpdateBalanceTx(tx, tt.accountID, tt.delta)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, tx.Rollback()) // Explicitly rollback the mock transaction
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestInsertTransactionLogTx tests the InsertTransactionLogTx method.
func TestPostgresAccountRepository_InsertTransactionLogTx(t *testing.T) {
	tests := []struct {
		name          string
		sourceID      int64
		destID        int64
		amount        float64
		mockExpect    func(sqlmock.Sqlmock)
		expectedTxID  string
		expectedError error
	}{
		{
			name:          "Successful transaction log insertion",
			sourceID:      100,
			destID:        200,
			amount:        50.00,
			mockExpect: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin() // Expect Begin for this transaction
				rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
				mock.ExpectQuery("INSERT INTO transactions").
					WithArgs(int64(100), int64(200), 50.00).
					WillReturnRows(rows)
				mock.ExpectRollback()
			},
			expectedTxID:  "1",
			expectedError: nil,
		},
		{
			name:          "Database error during transaction log insertion",
			sourceID:      101,
			destID:        201,
			amount:        75.00,
			mockExpect: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin() // Expect Begin for this transaction
				mock.ExpectQuery("INSERT INTO transactions").
					WithArgs(int64(101), int64(201), 75.00).
					WillReturnError(errors.New("tx log insert failed"))
				mock.ExpectRollback()
			},
			expectedTxID:  "",
			expectedError: errors.New("tx log insert failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupMockDB(t) // NEW: Get fresh mock DB per subtest
			repo := NewPostgresTransactionRepository(db) // NEW: Create repo with fresh DB

			tt.mockExpect(mock)
			
			tx, err := db.Begin() // Begin transaction on the fresh mock DB
			assert.NoError(t, err)

			txID, err := repo.InsertTransactionLogTx(tx, tt.sourceID, tt.destID, tt.amount)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTxID, txID)
			}
			assert.NoError(t, tx.Rollback())
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestIsSerializationFailure tests the IsSerializationFailure helper function.
func TestIsSerializationFailure(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "Serialization failure error",
			err:      errors.New("pq: deadlock detected (SQLSTATE 40001)"),
			expected: true,
		},
		{
			name:     "Another database error",
			err:      errors.New("pq: unique constraint violation (SQLSTATE 23505)"),
			expected: false,
		},
		{
			name:     "Non-database error",
			err:      errors.New("some generic error"),
			expected: false,
		},
		{
			name:     "Nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsSerializationFailure(tt.err))
		})
	}
}