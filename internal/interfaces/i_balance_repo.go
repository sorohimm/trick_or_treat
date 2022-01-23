package interfaces

import (
	"github.com/jackc/pgx/v4/pgxpool"
	"net/http"
	"users_balance/internal/models"
)

type ICompanyDetailsRepo interface {
	GetUserBalance(conn *pgxpool.Conn, uuid string) (models.User, error)
	UpdateAccount(conn *pgxpool.Conn, req models.UserBalanceUpdate) (models.User, error)
	CreateUser(conn *pgxpool.Conn, req models.UserBalanceUpdate) (models.User, error)
	InsertTransaction(conn *pgxpool.Conn, req models.UserBalanceUpdate) (models.Transaction, error)
	GetTransaction(conn *pgxpool.Conn, userUUID string, trxUUID string) (models.Transaction, error)
	GetTransactionsList(conn *pgxpool.Conn, userID string, limit int64, offset int64) ([]models.Transaction, error)
	GetExchangeRate(request *http.Request) (float64, error)
}
