package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"git.infra.egt.com/gateway/module/infrastructure"
)

// NewDatabasePool creates a new pgxpool.Pool.
func NewDatabasePool(log *zap.Logger, c *infrastructure.Config) (*pgxpool.Pool, error) {
	connString := c.DatabaseUrl

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		log.Error("Unable to create config", zap.Error(err))

		return nil, err
	}

	config.MaxConns = 5

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Error("Unable to create migrator", zap.Error(err))

		return nil, err
	}

	return pool, nil
}
