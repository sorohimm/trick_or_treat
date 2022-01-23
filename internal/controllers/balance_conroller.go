package balance_controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	er "users_balance/internal/errors"
	"users_balance/internal/interfaces"
	"users_balance/internal/models"
)

type UserBalanceController struct {
	Log                *zap.SugaredLogger
	UserBalanceService interfaces.IUserBalanceService
	Validator          *validator.Validate
}

func (c *UserBalanceController) GetUserBalance(ctx *gin.Context) {
	values := ctx.Request.URL.Query()

	request := models.User{
		ID:       values.Get("uuid"),
		Currency: values.Get("currency"),
	}

	if err := c.Validator.Struct(request); err != nil {
		c.Log.Info("validation : %s", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"message": er.ErrBadRequest.Error()})
		return
	}

	resp, err := c.UserBalanceService.GetUserBalance(request.ID, request.Currency)
	if err != nil {
		statusCode := ResolveErrorCode(err)
		c.Log.Infof(err.Error())
		ctx.JSON(statusCode, gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func (c *UserBalanceController) GetTransactionsList(ctx *gin.Context) {
	values := ctx.Request.URL.Query()

	limit, err := strconv.ParseInt(values.Get("limit"), 10, 64)
	offset, err := strconv.ParseInt(values.Get("offset"), 10, 54)
	request := models.TransactionsListRequest{
		UserID: values.Get("uuid"),
		Limit:  limit,
		Offset: offset,
		SortBy: values.Get("sort_by"),
		Cmp:    values.Get("cmp"),
	}

	if err := c.Validator.Struct(request); err != nil {
		c.Log.Info("validation : %s", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"message": er.ErrBadRequest.Error()})
		return
	}

	resp, err := c.UserBalanceService.GetTransactionsList(request)
	if err != nil {
		statusCode := ResolveErrorCode(err)
		c.Log.Infof(err.Error())
		ctx.JSON(statusCode, gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func (c *UserBalanceController) UpdateAccount(ctx *gin.Context) {
	var request models.UserBalanceUpdate

	err := ctx.BindJSON(&request)
	if err != nil {
		c.Log.Warn(err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "bad json :/"})
		return
	}

	if err := c.Validator.Struct(request); err != nil {
		c.Log.Infof("validation : %s", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"message": er.ErrBadRequest.Error()})
		return
	}

	resp, err := c.UserBalanceService.UpdateAccount(request)
	if err != nil {
		statusCode := ResolveErrorCode(err)
		c.Log.Infof(err.Error())
		ctx.JSON(statusCode, gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func (c *UserBalanceController) Transfer(ctx *gin.Context) {
	var request models.Transfer

	err := ctx.BindJSON(&request)
	if err != nil {
		c.Log.Warn(err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "bad json :/"})
		return
	}

	if err := c.Validator.Struct(request); err != nil {
		c.Log.Infof("validation : %s", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"message": er.ErrBadRequest.Error()})
		return
	}

	resp, err := c.UserBalanceService.Transfer(request)
	if err != nil {
		statusCode := ResolveErrorCode(err)
		c.Log.Infof(err.Error())
		ctx.JSON(statusCode, gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func ResolveErrorCode(err error) int {
	switch err {
	case er.ErrNotFound:
		return http.StatusNotFound
	case er.ErrInsufficientFunds:
		return http.StatusOK
	case er.ErrNegativeCreate:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
