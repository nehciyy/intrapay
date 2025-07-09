package service_test

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/nehciyy/intrapay/internal/service"
)
type MockAccountRepository struct {
	mock.Mock
}

func (m *MockAccountRepository) CreateAccount(accountID int64, initialBalance float64) error {
	args := m.Called(accountID, initialBalance)
	return args.Error(0)
}

func (m *MockAccountRepository) GetAccountBalance(accountID int64) (float64, error) {
	args := m.Called(accountID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockAccountRepository) AccountExists(accountID int64) (bool, error) {
	args := m.Called(accountID)
	return args.Bool(0), args.Error(1)
}

type MockTransactionRepository struct {
	mock.Mock
}

func (m *MockTransactionRepository) GetAccountBalanceTx(tx *sql.Tx, accountID int64) (float64, error) {
	args := m.Called(tx, accountID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockTransactionRepository) AccountExistsTx(tx *sql.Tx, accountID int64) (bool, error) {
	args := m.Called(tx, accountID)
	return args.Bool(0), args.Error(1)
}

func (m *MockTransactionRepository) UpdateBalanceTx(tx *sql.Tx, accountID int64, delta float64) error {
	args := m.Called(tx, accountID, delta)
	return args.Error(0)
}

func (m *MockTransactionRepository) InsertTransactionLogTx(tx *sql.Tx, sourceID, destID int64, amount float64) (string, error) {
	args := m.Called(tx, sourceID, destID, amount)
	return args.String(0), args.Error(1)
}

func newMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err, "failed to create mock db")
	t.Cleanup(func() { db.Close() })
	return db, mock
}

func TestCreateAccount(t *testing.T) {
	tests := []struct {
		name           string
		accountID      int64
		initialBalance float64
		mockExpect     func(*MockAccountRepository)
		expectedError  error
	}{
		{
			name:           "Success",
			accountID:      1,
			initialBalance: 100.0,
			mockExpect: func(mar *MockAccountRepository) {
				mar.On("CreateAccount", int64(1), 100.0).Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name:           "Duplicate Key Error",
			accountID:      1,
			initialBalance: 100.0,
			mockExpect: func(mar *MockAccountRepository) {
				mar.On("CreateAccount", int64(1), 100.0).Return(errors.New("duplicate key value violates unique constraint")).Once()
			},
			expectedError: errors.New("duplicate key value violates unique constraint"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, _ := newMockDB(t) 
			mockAccountRepo := new(MockAccountRepository)
			mockTransactionRepo := new(MockTransactionRepository) 

			svc := service.NewService(db, mockAccountRepo, mockTransactionRepo)

			tt.mockExpect(mockAccountRepo)

			err := svc.CreateAccount(tt.accountID, tt.initialBalance)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			mockAccountRepo.AssertExpectations(t)
		})
	}
}

func TestGetAccount(t *testing.T) {
	tests := []struct {
		name            string
		accountID       int64
		mockExpect      func(*MockAccountRepository)
		expectedBalance float64
		expectedError   error
	}{
		{
			name:      "Success",
			accountID: 1,
			mockExpect: func(mar *MockAccountRepository) {
				mar.On("GetAccountBalance", int64(1)).Return(250.5, nil).Once()
			},
			expectedBalance: 250.5,
			expectedError:   nil,
		},
		{
			name:      "Not Found",
			accountID: 1,
			mockExpect: func(mar *MockAccountRepository) {
				mar.On("GetAccountBalance", int64(1)).Return(float64(0), fmt.Errorf("account with ID %d not found", 1)).Once()
			},
			expectedBalance: 0,
			expectedError:   fmt.Errorf("account with ID %d not found", 1),
		},
		{
			name:      "Database Error",
			accountID: 1,
			mockExpect: func(mar *MockAccountRepository) {
				mar.On("GetAccountBalance", int64(1)).Return(float64(0), errors.New("db connection lost")).Once()
			},
			expectedBalance: 0,
			expectedError:   errors.New("db connection lost"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, _ := newMockDB(t)
			mockAccountRepo := new(MockAccountRepository)
			mockTransactionRepo := new(MockTransactionRepository)

			svc := service.NewService(db, mockAccountRepo, mockTransactionRepo)

			tt.mockExpect(mockAccountRepo)

			balance, err := svc.GetAccount(tt.accountID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBalance, balance)
			}
			mockAccountRepo.AssertExpectations(t)
		})
	}
}

func TestCreateTransaction(t *testing.T) {
	tests := []struct {
		name          string
		sourceID      int64
		destID        int64
		amount        float64
		mockExpect    func(*MockAccountRepository, *MockTransactionRepository) // No sqlmock.Sqlmock here
		expectedTxID  string
		expectedError error
		sqlMockExpect func(sqlmock.Sqlmock)
	}{
		{
			name:   "Success",
			sourceID: 1,
			destID: 2,
			amount: 100.0,
			mockExpect: func(mar *MockAccountRepository, mtr *MockTransactionRepository) {
				mtr.On("GetAccountBalanceTx", mock.Anything, int64(1)).Return(200.0, nil).Once()
				mtr.On("AccountExistsTx", mock.Anything, int64(2)).Return(true, nil).Once()
				mtr.On("UpdateBalanceTx", mock.Anything, int64(1), -100.0).Return(nil).Once()
				mtr.On("UpdateBalanceTx", mock.Anything, int64(2), 100.0).Return(nil).Once()
				mtr.On("InsertTransactionLogTx", mock.Anything, int64(1), int64(2), 100.0).Return("1234", nil).Once()
			},
			sqlMockExpect: func(mockDB sqlmock.Sqlmock) {
				mockDB.ExpectBegin()
				mockDB.ExpectCommit()
			},
			expectedTxID:  "1234",
			expectedError: nil,
		},
		{
			name:   "Insufficient Balance",
			sourceID: 1,
			destID: 2,
			amount: 100.0,
			mockExpect: func(mar *MockAccountRepository, mtr *MockTransactionRepository) {
				mtr.On("GetAccountBalanceTx", mock.Anything, int64(1)).Return(50.0, nil).Once() // Insufficient
			},
			sqlMockExpect: func(mockDB sqlmock.Sqlmock) {
				mockDB.ExpectBegin()
				mockDB.ExpectRollback()
			},
			expectedTxID:  "",
			expectedError: fmt.Errorf("insufficient balance in account %d", 1),
		},
		{
			name:   "Destination Account Not Found",
			sourceID: 1,
			destID: 2,
			amount: 100.0,
			mockExpect: func(mar *MockAccountRepository, mtr *MockTransactionRepository) {
				mtr.On("GetAccountBalanceTx", mock.Anything, int64(1)).Return(200.0, nil).Once()
				mtr.On("AccountExistsTx", mock.Anything, int64(2)).Return(false, nil).Once() // Not found
			},
			sqlMockExpect: func(mockDB sqlmock.Sqlmock) {
				mockDB.ExpectBegin()
				mockDB.ExpectRollback()
			},
			expectedTxID:  "",
			expectedError: fmt.Errorf("destination account %d not found", 2),
		},
		{
			name:   "Max Retries Exceeded",
			sourceID: 1,
			destID: 2,
			amount: 10.0,
			mockExpect: func(mar *MockAccountRepository, mtr *MockTransactionRepository) {
				for i := 0; i < 3; i++ {
					mtr.On("GetAccountBalanceTx", mock.Anything, int64(1)).Return(2000.0, nil).Once()
					mtr.On("AccountExistsTx", mock.Anything, int64(2)).Return(true, nil).Once()
					mtr.On("UpdateBalanceTx", mock.Anything, int64(1), -10.0).Return(nil).Once()
					mtr.On("UpdateBalanceTx", mock.Anything, int64(2), 10.0).Return(nil).Once()
					mtr.On("InsertTransactionLogTx", mock.Anything, int64(1), int64(2), 10.0).Return("temp_id", nil).Once()
				}
			},
			sqlMockExpect: func(mockDB sqlmock.Sqlmock) {
				// Simulate DB calls for all 3 retries
				for i := 0; i < 3; i++ {
					mockDB.ExpectBegin()
					mockDB.ExpectCommit().WillReturnError(fmt.Errorf("pq: deadlock detected (SQLSTATE 40001)"))
				}
			},
			expectedTxID:  "",
			expectedError: errors.New("transaction failed after max retries"),
		},
		{
			name:   "Begin Transaction Failure",
			sourceID: 1,
			destID: 2,
			amount: 100.0,
			mockExpect: func(mar *MockAccountRepository, mtr *MockTransactionRepository) {
				// No repository mocks needed as Begin fails immediately
			},
			sqlMockExpect: func(mockDB sqlmock.Sqlmock) {
				mockDB.ExpectBegin().WillReturnError(errors.New("failed to connect to db"))
			},
			expectedTxID:  "",
			expectedError: errors.New("failed to begin transaction: failed to connect to db"),
		},
		{
			name:   "Commit Failure (Non-Serialization)",
			sourceID: 1,
			destID: 2,
			amount: 100.0,
			mockExpect: func(mar *MockAccountRepository, mtr *MockTransactionRepository) {
				mtr.On("GetAccountBalanceTx", mock.Anything, int64(1)).Return(200.0, nil).Once()
				mtr.On("AccountExistsTx", mock.Anything, int64(2)).Return(true, nil).Once()
				mtr.On("UpdateBalanceTx", mock.Anything, int64(1), -100.0).Return(nil).Once()
				mtr.On("UpdateBalanceTx", mock.Anything, int64(2), 100.0).Return(nil).Once()
				mtr.On("InsertTransactionLogTx", mock.Anything, int64(1), int64(2), 100.0).Return("some-id", nil).Once()
			},
			sqlMockExpect: func(mockDB sqlmock.Sqlmock) {
				mockDB.ExpectBegin()
				mockDB.ExpectCommit().WillReturnError(errors.New("network error on commit"))
			},
			expectedTxID:  "",
			expectedError: errors.New("commit failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mockDB := newMockDB(t) // Fresh mock DB for each subtest
			mockAccountRepo := new(MockAccountRepository)
			mockTransactionRepo := new(MockTransactionRepository)

			// Set sqlmock expectations for Begin/Commit/Rollback for this specific test case
			tt.sqlMockExpect(mockDB)
			// Set testify/mock expectations for repository methods
			tt.mockExpect(mockAccountRepo, mockTransactionRepo)

			svc := service.NewService(db, mockAccountRepo, mockTransactionRepo)

			id, err := svc.CreateTransaction(tt.sourceID, tt.destID, tt.amount)
			if tt.expectedError != nil {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.expectedError.Error())
				require.Equal(t, tt.expectedTxID, id)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedTxID, id)
			}


			// Verify all expectations for both sqlmock and testify/mock
			assert.NoError(t, mockDB.ExpectationsWereMet(), "sqlmock expectations not met")
			mockAccountRepo.AssertExpectations(t)
			mockTransactionRepo.AssertExpectations(t)
		})
	}
}