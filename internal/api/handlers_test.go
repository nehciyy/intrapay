package api_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/nehciyy/intrapay/internal/api"
	"github.com/nehciyy/intrapay/internal/models"
)

type mockService struct {
	CreateAccountFn     func(id int64, balance float64) error
	GetAccountFn        func(id int64) (float64, error)
	CreateTransactionFn func(from, to int64, amount float64) (string, error)
}

func (m *mockService) CreateAccount(id int64, balance float64) error {
	return m.CreateAccountFn(id, balance)
}

func (m *mockService) GetAccount(id int64) (float64, error) {
	return m.GetAccountFn(id)
}

func (m *mockService) CreateTransaction(from, to int64, amount float64) (string, error) {
	return m.CreateTransactionFn(from, to, amount)
}


// --- CreateAccount Tests ---
func TestCreateAccount_Success(t *testing.T) {
	server := &api.Server{
		Service: &mockService{
			CreateAccountFn: func(id int64, balance float64) error {
				return nil
			},
		},
	}
	body := models.CreateAccountRequest{AccountID: 123, InitialBalance: 100.0}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/accounts", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	server.CreateAccount(resp, req)

	if resp.Code != http.StatusCreated {
		t.Errorf("expected status 201 Created, got %d", resp.Code)
	}
}

func TestCreateAccount_InvalidJSON(t *testing.T) {
	server := &api.Server{Service: &mockService{}}
	req := httptest.NewRequest("POST", "/accounts", strings.NewReader("invalid json"))
	resp := httptest.NewRecorder()

	server.CreateAccount(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.Code)
	}
}

// --- GetAccount Tests ---

func TestGetAccount_Success(t *testing.T) {
	server := &api.Server{
		Service: &mockService{
			GetAccountFn: func(id int64) (float64, error) {
				return 200.50, nil
			},
		},
	}

	req := httptest.NewRequest("GET", "/accounts/123", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/accounts/{id}", server.GetAccount)
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp["account_id"] != float64(123) || resp["balance"] != 200.50 {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestGetAccount_InvalidID(t *testing.T) {
	server := &api.Server{Service: &mockService{}}
	req := httptest.NewRequest("GET", "/accounts/abc", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/accounts/{id}", server.GetAccount)
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestGetAccount_NotFound(t *testing.T) {
	server := &api.Server{
		Service: &mockService{
			GetAccountFn: func(id int64) (float64, error) {
				return 0, errors.New("not found")
			},
		},
	}

	req := httptest.NewRequest("GET", "/accounts/999", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/accounts/{id}", server.GetAccount)
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}
// --- CreateTransaction Tests ---

func TestCreateTransaction_Success(t *testing.T) {
	server := &api.Server{
		Service: &mockService{
			CreateTransactionFn: func(from, to int64, amount float64) (string, error) {
				return "tx123", nil
			},
		},
	}

	reqBody := models.TransactionRequest{
		SourceAccountID:      1,
		DestinationAccountID: 2,
		Amount:               50.0,
	}
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/transactions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	server.CreateTransaction(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rr.Code)
	}

	var resp map[string]string
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp["transaction_id"] != "tx123" {
		t.Errorf("expected tx123, got %s", resp["transaction_id"])
	}
}


func TestCreateTransaction_InvalidJSON(t *testing.T) {
	server := &api.Server{Service: &mockService{}}
	req := httptest.NewRequest("POST", "/transactions", strings.NewReader("invalid"))
	rr := httptest.NewRecorder()

	server.CreateTransaction(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestCreateTransaction_Failure(t *testing.T) {
	server := &api.Server{
		Service: &mockService{
			CreateTransactionFn: func(from, to int64, amount float64) (string, error) {
				return "", errors.New("failed to process transaction")
			},
		},
	}

	reqBody := models.TransactionRequest{
		SourceAccountID:      1,
		DestinationAccountID: 2,
		Amount:               50.0,
	}
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/transactions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	server.CreateTransaction(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}