package service

type Service interface {
	CreateAccount(accountID int64, initialBalance float64) error
	GetAccount(accountID int64) (float64, error)
	CreateTransaction(sourceID int64, destID int64, amount float64) (string, error)
}
