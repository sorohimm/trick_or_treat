package balance_services

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/http"
	"sort"
	"users_balance/internal/config"
	er "users_balance/internal/errors"
	"users_balance/internal/interfaces"
	"users_balance/internal/models"
)

type UserBalanceService struct {
	Log         *zap.SugaredLogger
	Config      *config.Config
	BalanceRepo interfaces.ICompanyDetailsRepo
	DBHandler   interfaces.IDBHandler
}

//provide users balance
func (s *UserBalanceService) GetUserBalance(uuid string, currency string) (models.User, error) {
	conn, err := s.DBHandler.AcquireConn(context.Background())
	if err != nil {
		s.Log.Info(err.Error())
		return models.User{}, err
	}
	defer conn.Release()

	result, err := s.BalanceRepo.GetUserBalance(conn, uuid)
	switch {
	case errors.Cause(err) == pgx.ErrNoRows:
		return models.User{}, er.ErrNotFound
	case err != nil:
		return models.User{}, err
	}

	if currency != "" && currency != "RUB" {
		s.calculateExchangeBalance(&result, currency)
		return result, nil
	}

	result.Currency = models.RUB
	return result, nil
}

func (s *UserBalanceService) UpdateAccount(req models.UserBalanceUpdate) (models.UserBalanceUpdateResponse, error) {
	conn, err := s.DBHandler.AcquireConn(context.Background())
	if err != nil {
		return models.UserBalanceUpdateResponse{}, err
	}
	defer conn.Release()

	if req.Amount < 0 {
		result, err := s.BalanceRepo.GetUserBalance(conn, req.UserID)
		switch {
		case err == nil && result.Balance+req.Amount < 0:
			return models.UserBalanceUpdateResponse{}, er.ErrInsufficientFunds
		case errors.Cause(err) == pgx.ErrNoRows:
			return models.UserBalanceUpdateResponse{}, er.ErrNegativeCreate
		}
	}

	user, err := s.BalanceRepo.UpdateAccount(conn, req)
	switch {
	case errors.Cause(err) == pgx.ErrNoRows:
		result, err := s.createNewUser(conn, req)
		if err != nil {
			return models.UserBalanceUpdateResponse{}, err
		}
		return result, nil
	case err != nil:
		return models.UserBalanceUpdateResponse{}, err
	}

	trx, err := s.BalanceRepo.InsertTransaction(conn, req)
	if err != nil {
		return models.UserBalanceUpdateResponse{}, err
	}

	result := models.UserBalanceUpdateResponse{
		User:        user,
		Transaction: trx,
	}

	return result, nil
}

func (s *UserBalanceService) Transfer(req models.Transfer) (models.TransferResponse, error) {
	conn, err := s.DBHandler.AcquireConn(context.Background())
	if err != nil {
		s.Log.Info("acquire conn error")
		return models.TransferResponse{}, err
	}
	defer conn.Release()

	err = s.isTransferPossible(conn, req.From, req.To, req.Amount)
	if err != nil {
		return models.TransferResponse{}, err
	}

	res, err := s.doTransfer(conn, req)
	if err != nil {
		return models.TransferResponse{}, err
	}

	return res, nil
}

func (s *UserBalanceService) GetTransactionsList(req models.TransactionsListRequest) (models.TransactionsListResponse, error) {
	conn, err := s.DBHandler.AcquireConn(context.Background())
	if err != nil {
		s.Log.Info("acquire conn error")
		return models.TransactionsListResponse{}, err
	}
	defer conn.Release()

	list, err := s.BalanceRepo.GetTransactionsList(conn, req.UserID, req.Limit, req.Offset)
	switch {
	case errors.Cause(err) == pgx.ErrNoRows:
		return models.TransactionsListResponse{}, er.ErrNotFound
	case err != nil:
		return models.TransactionsListResponse{}, err
	case len(list) == 0:
		return models.TransactionsListResponse{}, er.ErrNotFound
	}

	trxsort(list, req.SortBy, req.Cmp)

	result := models.TransactionsListResponse{
		TransactionsList: list,
	}

	return result, nil
}

func trxsort(list []models.Transaction, by string, cmp string) {
	switch by {
	case "date":
		if cmp == "d" {
			sort.SliceStable(list, func(i, j int) bool {
				return list[i].Timestamp > list[j].Timestamp
			})
		} else if cmp == "i" {
			sort.SliceStable(list, func(i, j int) bool {
				return list[i].Timestamp < list[j].Timestamp
			})
		}
	case "amount":
		if cmp == "d" {
			sort.SliceStable(list, func(i, j int) bool {
				return list[i].Amount > list[j].Amount
			})
		} else if cmp == "i" {
			sort.SliceStable(list, func(i, j int) bool {
				return list[i].Amount < list[j].Amount
			})
		}
	}
}

func (s *UserBalanceService) createNewUser(conn *pgxpool.Conn, req models.UserBalanceUpdate) (models.UserBalanceUpdateResponse, error) {
	user, err := s.BalanceRepo.CreateUser(conn, req)
	if err != nil {
		return models.UserBalanceUpdateResponse{}, err
	}
	trx, err := s.BalanceRepo.InsertTransaction(conn, req)
	if err != nil {
		return models.UserBalanceUpdateResponse{}, err
	}

	result := models.UserBalanceUpdateResponse{
		User:        user,
		Transaction: trx,
	}

	return result, nil
}

func (s *UserBalanceService) doTransfer(conn *pgxpool.Conn, req models.Transfer) (models.TransferResponse, error) {
	const senderTransferDescriptionStatement = `transfer to another user`
	sender := models.UserBalanceUpdate{
		UserID:      req.From,
		Who:         req.From,
		Description: senderTransferDescriptionStatement,
		Amount:      -req.Amount,
		Currency:    models.RUB,
	}
	senderUpd, err := s.UpdateAccount(sender)
	if err != nil {
		return models.TransferResponse{}, err
	}

	const recipientTransferDescriptionStatement = `transfer from another user`
	recipient := models.UserBalanceUpdate{
		UserID:      req.To,
		Who:         req.From,
		Description: recipientTransferDescriptionStatement,
		Amount:      req.Amount,
		Currency:    models.RUB,
	}
	_, err = s.UpdateAccount(recipient)
	if err != nil {
		s.abortTransaction(conn, sender.UserID, senderUpd.Transaction.TrxID)
		return models.TransferResponse{}, err
	}

	return models.TransferResponse{Success: "true"}, nil
}

func (s *UserBalanceService) abortTransaction(conn *pgxpool.Conn, userUUID string, trxUUID string) bool {
	trx, err := s.BalanceRepo.GetTransaction(conn, userUUID, trxUUID)
	if err != nil {
		return false
	}

	abortSatement := models.UserBalanceUpdate{
		UserID:      userUUID,
		Who:         "server",
		Description: fmt.Sprintf("abort transaction %s", trxUUID),
		Amount:      -trx.Amount,
		Currency:    trx.Currency,
	}
	_, err = s.BalanceRepo.UpdateAccount(conn, abortSatement)
	if err != nil {
		return false
	}

	_, err = s.BalanceRepo.InsertTransaction(conn, abortSatement)
	if err != nil {
		return false
	}

	return true
}

func (s *UserBalanceService) isTransferPossible(conn *pgxpool.Conn, senderUUID string, recipientUUID string, amount float64) error {
	sender, err := s.BalanceRepo.GetUserBalance(conn, senderUUID)
	_, err = s.BalanceRepo.GetUserBalance(conn, recipientUUID)
	switch {
	case errors.Cause(err) == pgx.ErrNoRows:
		return er.ErrNotFound
	case err != nil:
		return err
	case sender.Balance >= amount:
		return nil
	default:
		return er.ErrNegativeBalance
	}
}

func (s *UserBalanceService) calculateExchangeBalance(v *models.User, currency string) {
	req, _ := http.NewRequest(http.MethodGet, s.Config.APIData.URL, nil)
	req.URL.Path = s.Config.APIData.Path
	q := req.URL.Query()
	q.Add("apikey", s.Config.Key)
	q.Add("base_currency", currency)
	req.URL.RawQuery = q.Encode()

	ExchangeAmount, err := s.BalanceRepo.GetExchangeRate(req)
	if err != nil {
		v.Currency = models.RUB
		return
	}

	v.Balance = v.Balance / ExchangeAmount
	v.Currency = currency
}
