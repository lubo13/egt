package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type Message struct {
	ID        int64             `db:"id"`
	Topic     string            `db:"topic"`
	Headers   map[string]string `db:"headers"`
	Message   []byte            `db:"message"`
	CreatedAt time.Time         `db:"created_at"`
}

type MessageRepository struct {
	queryExecutor QueryExecutor
}

// NewMessageRepository receives the pool from Fx.
func NewMessageRepository(queryExecutor QueryExecutor) *MessageRepository {
	return &MessageRepository{queryExecutor: queryExecutor}
}

func (r *MessageRepository) FetchBatch(
	ctx context.Context,
) ([]*Message, error) {
	sql := `SELECT * FROM messages order by id limit 100`

	rows, err := r.queryExecutor.Query(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Message])
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *MessageRepository) Delete(
	ctx context.Context,
	ids []int64,
) error {
	sql := `DELETE FROM messages WHERE id=ANY($1)`

	_, err := r.queryExecutor.Exec(ctx, sql, ids)
	if err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}

	return nil
}
