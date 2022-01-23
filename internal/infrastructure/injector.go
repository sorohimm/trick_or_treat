package infrastructure

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"net/http"
	"users_balance/internal/config"
	"users_balance/internal/controllers"
	"users_balance/internal/interfaces"
	"users_balance/internal/repos"
	"users_balance/internal/services"
)

type IInjector interface {
	InjectBalanceController() balance_controllers.UserBalanceController
}

var env *environment

type environment struct {
	logger   *zap.SugaredLogger
	cfg      *config.Config
	client   *http.Client
	dbClient interfaces.IDBHandler
}

func (e *environment) InjectBalanceController() balance_controllers.UserBalanceController {
	return balance_controllers.UserBalanceController{
		Log: e.logger,
		UserBalanceService: &balance_services.UserBalanceService{
			Log: e.logger,
			BalanceRepo: &balance_repos.UserBalanceRepo{
				Log:    e.logger,
				Client: http.DefaultClient,
				Config: e.cfg,
			},
			Config:    e.cfg,
			DBHandler: e.dbClient,
		},
		Validator: validator.New(),
	}
}

func Injector(log *zap.SugaredLogger, cfg *config.Config) (IInjector, error) {
	client, err := InitPostgresClient(cfg)
	if err != nil {
		log.Fatal("injector :: db init error")
		return nil, err
	}

	env = &environment{
		logger:   log,
		cfg:      cfg,
		client:   http.DefaultClient,
		dbClient: client,
	}

	return env, nil
}
