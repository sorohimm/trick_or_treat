package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"os"
	"users_balance/internal/config"
	"users_balance/internal/infrastructure"
)

var (
	cfg *config.Config
	ctx context.Context
	log *zap.SugaredLogger
)

func init() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Printf("error loading logger: %s", err)
		os.Exit(1)
		return
	}

	log = logger.Sugar()

	cfg, err = config.New()
	if err != nil {
		log.Fatalf("config init error :: %s", err)
	}
	log.Infof("config loaded ::\n%+v", cfg)
}

func main() {
	injector, err := infrastructure.Injector(log, cfg)
	if err != nil {
		log.Fatal("main :: inject failing")
	}

	balanceController := injector.InjectBalanceController()

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	v1 := router.Group("/cash/v1")
	{
		v1.GET("/balance", balanceController.GetUserBalance)
		v1.POST("/balance/update", balanceController.UpdateAccount)
		v1.POST("/balance/transfer", balanceController.Transfer)
		v1.GET("/trx_list", balanceController.GetTransactionsList)
	}

	err = router.Run()
	if err != nil {
		log.Fatal("main :: router start error")
	}
}
