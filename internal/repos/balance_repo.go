package balance_repos

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
	"net/http"
	"users_balance/internal/config"
	"users_balance/internal/models"
)

type UserBalanceRepo struct {
	Log    *zap.SugaredLogger
	Client *http.Client
	Config *config.Config
}

func (r *UserBalanceRepo) GetUserBalance(conn *pgxpool.Conn, uuid string) (models.User, error) {
	const GetUserBalanceStatement = `SELECT * FROM users WHERE uuid = $1;`
	var user models.User
	err := conn.QueryRow(context.Background(), GetUserBalanceStatement, uuid).Scan(&user.ID, &user.Balance)
	if err != nil {
		r.Log.Info(err.Error())
		return models.User{}, err
	}

	return user, nil
}

func (r *UserBalanceRepo) UpdateAccount(conn *pgxpool.Conn, req models.UserBalanceUpdate) (models.User, error) {
	const UpdateAccountStatement = `UPDATE users SET balance = balance + $2 WHERE uuid = $1
								   RETURNING "uuid", "balance";`

	var user models.User
	err := conn.QueryRow(context.Background(), UpdateAccountStatement, req.UserID, req.Amount).Scan(&user.ID, &user.Balance)
	if err != nil {
		r.Log.Info(err.Error())
		return models.User{}, err
	}

	return user, nil
}

func (r *UserBalanceRepo) CreateUser(conn *pgxpool.Conn, req models.UserBalanceUpdate) (models.User, error) {
	const CreateUserStatement = `INSERT INTO users (uuid, balance) VALUES ($1, $2) 
								 RETURNING "uuid", "balance";`

	var user models.User
	err := conn.QueryRow(context.Background(), CreateUserStatement, req.UserID, req.Amount).Scan(&user.ID, &user.Balance)

	if err != nil {
		r.Log.Info(err.Error())
		return models.User{}, err
	}

	return user, nil
}

func (r *UserBalanceRepo) InsertTransaction(conn *pgxpool.Conn, req models.UserBalanceUpdate) (models.Transaction, error) {
	const UpdateTransactionListStatement = `INSERT INTO transactions (user_uuid, who, description, amount, currency) 
											VALUES ($1, $2, $3, $4, $5)
											RETURNING "trx_uuid", CAST("trx_date" AS text), CAST("trx_time" AS text);`

	trx := models.Transaction{
		Who:         req.Who,
		Description: req.Description,
		Amount:      req.Amount,
		Currency:    req.Currency,
	}

	err := conn.QueryRow(context.Background(), UpdateTransactionListStatement, req.UserID, req.Who, req.Description,
		req.Amount, req.Currency).Scan(&trx.TrxID, &trx.Date, &trx.Time)
	if err != nil {
		r.Log.Info(err.Error())
		return models.Transaction{}, err
	}

	return trx, nil
}

func (r *UserBalanceRepo) GetTransaction(conn *pgxpool.Conn, userUUID string, trxUUID string) (models.Transaction, error) {
	const GetTransactionStatement = `SELECT trx_uuid, CAST("trx_date" AS text), CAST("trx_time" AS text), who, description, amount, currency 
									  FROM transactions WHERE uuid = $1 && rtx_uuid = $2;`

	var trx models.Transaction
	err := conn.QueryRow(context.Background(), GetTransactionStatement, userUUID,
		trxUUID).Scan(&trx.TrxID, &trx.Date, &trx.Time, &trx.Who, &trx.Description, &trx.Amount, &trx.Currency)
	if err != nil {
		r.Log.Info(err.Error())
		return models.Transaction{}, err
	}

	return trx, nil
}

func (r *UserBalanceRepo) GetTransactionsList(conn *pgxpool.Conn, userID string, limit int64, offset int64) ([]models.Transaction, error) {
	const GetTransactionsListStatement = `SELECT trx_uuid, CAST("trx_date" AS text), CAST("trx_time" AS text), u_timestamp, who, description, amount, currency 
									  	   FROM transactions WHERE user_uuid = $1 
									  	   LIMIT $2
									 	   OFFSET $3;`

	var trxList []models.Transaction
	rows, err := conn.Query(context.Background(), GetTransactionsListStatement, userID, limit, offset)
	if err != nil {
		r.Log.Info(err.Error())
		return nil, err
	}

	for rows.Next() {
		var trx models.Transaction
		err := rows.Scan(&trx.TrxID, &trx.Date, &trx.Time, &trx.Timestamp, &trx.Who, &trx.Description, &trx.Amount, &trx.Currency)
		if err != nil {
			r.Log.Info(err.Error())
			return nil, err
		}
		trxList = append(trxList, trx)
	}

	return trxList, nil
}

func (r *UserBalanceRepo) GetExchangeRate(request *http.Request) (float64, error) {
	resp, err := r.Client.Do(request)
	if err != nil || resp.StatusCode != http.StatusOK {
		return 0, err
	}

	result := models.Exchange{}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&result)
	if err != nil {
		r.Log.Info("decode error")
		return 0, err
	}

	return result.Data[models.RUB], nil
}
