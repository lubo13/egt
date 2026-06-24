package database

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const DB_TX string = "db.tx"

type TransactionManager struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

func NewTransactionManager(pool *pgxpool.Pool, logger *zap.Logger) *TransactionManager {
	return &TransactionManager{
		pool:   pool,
		logger: logger,
	}
}

func (m *TransactionManager) WrapInTransaction(ctx context.Context, c func(ctx context.Context) error) error {
	tx := extractTransactionFromContext(ctx)
	var err error
	isExist := true
	if tx == nil {
		isExist = false

		tx, err = m.pool.Begin(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback(ctx)

		m.logger.Info("new trasaction begin")

		ctx = putTransactionToContext(ctx, tx)
	}

	err = c(ctx)
	if err != nil {
		return err
	}

	if !isExist {
		m.logger.Info("trasaction commit")

		return tx.Commit(ctx)
	}

	return nil
}

func (m *TransactionManager) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	tx := extractTransactionFromContext(ctx)
	if tx == nil {
		m.logger.Info("new connection aquired")

		return m.pool.Exec(ctx, sql, args...)
	}

	m.logger.Info("trasaction reused")

	return tx.Exec(ctx, sql, args...)
}

func (m *TransactionManager) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	tx := extractTransactionFromContext(ctx)
	if tx == nil {
		m.logger.Info("new connection aquired")

		return m.pool.Query(ctx, sql, args...)
	}

	m.logger.Info("trasaction reused")

	return tx.Query(ctx, sql, args...)
}

// TODO this is anti-pattern, but I want to avoid coupling application layer with infrastructure (transactions)
func extractTransactionFromContext(ctx context.Context) pgx.Tx {
	if tx, ok := ctx.Value(DB_TX).(pgx.Tx); ok {
		return tx
	}

	return nil
}

func putTransactionToContext(
	ctx context.Context,
	tx pgx.Tx,
) context.Context {
	return context.WithValue(ctx, DB_TX, tx)
}
