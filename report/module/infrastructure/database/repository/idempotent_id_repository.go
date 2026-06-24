package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type IdempotentIdRepository struct {
	queryExecutor QueryExecutorInterface
	logger        *zap.Logger
}

// NewIdempotentIdRepository receives the queryExecutor from Fx.
func NewIdempotentIdRepository(queryExecutor QueryExecutorInterface, logger *zap.Logger) *IdempotentIdRepository {
	return &IdempotentIdRepository{queryExecutor: queryExecutor, logger: logger}
}

func (r *IdempotentIdRepository) IsExists(ctx context.Context, id *uuid.UUID) (bool, error) {
	sql := `SELECT EXISTS(SELECT 1 FROM idempotent_ids WHERE id = $1)`

	var isExists bool
	err := r.queryExecutor.QueryRow(ctx, sql, id).Scan(&isExists)
	if err != nil {
		return false, err
	}

	return isExists, nil
}

func (r *IdempotentIdRepository) Save(ctx context.Context, id *uuid.UUID) error {
	sql := `INSERT INTO idempotent_ids (id, created_at)
				VALUES ($1, $2)`

	_, err := r.queryExecutor.Exec(
		ctx,
		sql,
		id,
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("insert failed: %w", err)
	}

	return nil
}
