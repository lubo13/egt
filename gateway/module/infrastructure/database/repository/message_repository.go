package repository

import (
	"context"
	"fmt"
	"time"
)

type MessageRepository struct {
	queryExecutor QueryExecutor
}

// NewMessageRepository receives the pool from Fx.
func NewMessageRepository(queryExecutor QueryExecutor) *MessageRepository {
	return &MessageRepository{queryExecutor: queryExecutor}
}

func (r *MessageRepository) Produce(
	ctx context.Context,
	topic string,
	headers []byte,
	message []byte,
	createdAt time.Time,
) error {
	sql := `INSERT INTO messages (topic, headers, message, created_at) VALUES ($1, $2, $3, $4)`

	_, err := r.queryExecutor.Exec(ctx, sql, topic, headers, message, createdAt)
	if err != nil {
		return fmt.Errorf("insert failed: %w", err)
	}

	return nil
}
