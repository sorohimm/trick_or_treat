package models

type Transfer struct {
	From   string  `json:"from" validate:"required,uuid"`
	To     string  `json:"to" validate:"required,uuid"`
	Amount float64 `json:"amount" validate:"gt=0"`
}

type User struct {
	ID       string  `json:"uuid" validate:"required,uuid"`
	Balance  float64 `json:"balance" validate:"omitempty"`
	Currency string  `json:"currency,omitempty" validate:"omitempty"`
}

type UserBalanceUpdate struct {
	UserID      string  `json:"uuid" validate:"required,uuid"`
	Who         string  `json:"who" validate:"required"`
	Description string  `json:"description" validate:"omitempty"`
	Amount      float64 `json:"amount" validate:"required"`
	Currency    string  `json:"currency" validate:"required"`
}

type UserBalanceUpdateResponse struct {
	User        User        `json:"user"`
	Transaction Transaction `json:"transaction"`
}

type Transaction struct {
	TrxID       string  `json:"id"`
	Date        string  `json:"date"`
	Time        string  `json:"time"`
	Timestamp   int     `json:"timestamp"`
	Who         string  `json:"who"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
}

type TransferResponse struct {
	Success string `json:"success,required"`
}

type TransactionsListRequest struct {
	UserID string `json:"uuid" validate:"required,uuid"`
	Limit  int64  `json:"limit" validate:"required,gte=10,lte=100"`
	Offset int64  `json:"offset" validate:"omitempty,gte=0"`
	SortBy string `json:"sort_by" validate:"omitempty,oneof=date amount"`
	Cmp    string `json:"cmp" validate:"omitempty,oneof=d i"`
}

type TransactionsListResponse struct {
	TransactionsList []Transaction `json:"transactions,omitempty"`
}

type Exchange struct {
	Data map[string]float64 `json:"data,omitempty"`
}

const (
	RUB = "RUB"
)
