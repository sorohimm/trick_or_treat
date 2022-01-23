package infrastructure

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"log"
	"users_balance/internal/config"
	"users_balance/internal/interfaces"
)

type PostgresClient struct {
	Pool *pgxpool.Pool
}

func InitPostgresClient(cfg *config.Config) (interfaces.IDBHandler, error) {
	pool, err := pgxpool.Connect(context.Background(), fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBAdminUsername, cfg.DBAdminPassword, cfg.DBHost, cfg.DBPort, cfg.DBName))
	if err != nil {
		log.Print(err)
		return nil, errors.Wrap(err, "postgres init")
	}

	return &PostgresClient{Pool: pool}, nil
}

func (p *PostgresClient) GetPool() *pgxpool.Pool {
	return p.Pool
}

func (p *PostgresClient) AcquireConn(ctx context.Context) (*pgxpool.Conn, error) {
	return p.Pool.Acquire(ctx)
}

func (p *PostgresClient) StartTransaction(ctx context.Context) (pgx.Tx, error) {
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Begin")
	}

	return tx, err
}

func (p *PostgresClient) FinishTransaction(ctx context.Context, tx pgx.Tx, err error) error {
	if err != nil {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			return errors.Wrap(err, "Rollback")
		}

		return err
	} else {
		if commitErr := tx.Commit(ctx); commitErr != nil {
			return errors.Wrap(err, "failed to commit tx")
		}

		return nil
	}
}
