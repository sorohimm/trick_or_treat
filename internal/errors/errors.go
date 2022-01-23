package errors

import "errors"

var ErrNotFound = errors.New("user not found")
var ErrInsufficientFunds = errors.New("insufficient funds")
var ErrBadRequest = errors.New("bad request :/")
var ErrNegativeCreate = errors.New("the user does not exist, it is impossible to create a user with a negative balance")
var ErrNegativeBalance = errors.New("transfer is prohibited, insufficient funds")