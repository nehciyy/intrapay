package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/nehciyy/intrapay/internal/models"
	"github.com/nehciyy/intrapay/internal/service"
)

type Server struct {
	DB *sql.DB
	Service service.Service
}

func (s *Server) CreateAccount(w http.ResponseWriter, r *http.Request) {
	req := &models.CreateAccountRequest{}

	if err := json.NewDecoder(r.Body).Decode(req); err != nil{
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.Service.CreateAccount(s.DB, req.AccountID, req.InitialBalance); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
    
    w.WriteHeader(http.StatusCreated)
}

func (s *Server) GetAccount(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
    if err != nil {
        http.Error(w, "invalid account ID", http.StatusBadRequest)
        return
    }

    balance, err := s.Service.GetAccount(s.DB, id)
	if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    json.NewEncoder(w).Encode(map[string]interface{}{
        "account_id": id,
        "balance":    balance,
    })
}

func (s *Server) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	req := &models.TransactionRequest{}

	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	transactionID, err := s.Service.CreateTransaction(s.DB, req.SourceAccountID, req.DestinationAccountID, req.Amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Transaction successfully processed",
		"transaction_id": transactionID,
	})

}