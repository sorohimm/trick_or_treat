package interfaces

import (
	"users_balance/internal/models"
)

type IUserBalanceService interface {
	GetUserBalance(string, string) (models.User, error)
	UpdateAccount(models.UserBalanceUpdate) (models.UserBalanceUpdateResponse, error)
	Transfer(req models.Transfer) (models.TransferResponse, error)
	GetTransactionsList(req models.TransactionsListRequest) (models.TransactionsListResponse, error)
}
