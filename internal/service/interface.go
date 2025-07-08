package service

import "database/sql"

type Service interface {
	CreateAccount(db *sql.DB, id int64, balance float64) error
	GetAccount(db *sql.DB, id int64) (float64, error)
	CreateTransaction(db *sql.DB, from, to int64, amount float64) (string, error)
}