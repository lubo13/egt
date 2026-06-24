package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	"git.infra.egt.com/report/module/internal/domain"
)

type DeviceEventRepository struct {
	queryExecutor QueryExecutorInterface
	logger        *zap.Logger
}

// NewDeviceEventRepository receives the queryExecutor from Fx.
func NewDeviceEventRepository(queryExecutor QueryExecutorInterface, logger *zap.Logger) *DeviceEventRepository {
	return &DeviceEventRepository{queryExecutor: queryExecutor, logger: logger}
}

func (r *DeviceEventRepository) FetchAll(ctx context.Context) ([]*domain.DeviceEvent, error) {
	sql := `SELECT * FROM device_events ORDER BY created_at ASC`

	rows, err := r.queryExecutor.Query(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	deviceEvents, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[domain.DeviceEvent])
	if err != nil {
		return nil, err
	}

	return deviceEvents, nil
}

func (r *DeviceEventRepository) FetchByID(ctx context.Context, id *uuid.UUID) (*domain.DeviceEvent, error) {
	sql := `SELECT * FROM device_events WHERE id = $1`

	deviceEvent := &domain.DeviceEvent{}
	row := r.queryExecutor.QueryRow(ctx, sql, id.String())

	err := row.Scan(
		&deviceEvent.ID,
		&deviceEvent.DeviceId,
		&deviceEvent.DeviceName,
		&deviceEvent.SensorName,
		&deviceEvent.Message,
		&deviceEvent.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return deviceEvent, nil
}

func (r *DeviceEventRepository) Save(ctx context.Context, deviceEvent *domain.DeviceEvent) error {
	sql := `INSERT INTO device_events (id, device_id, device_name, sensor_name, message, created_at)
				VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.queryExecutor.Exec(
		ctx,
		sql,
		deviceEvent.ID,
		deviceEvent.DeviceId,
		deviceEvent.DeviceName,
		deviceEvent.SensorName,
		deviceEvent.Message,
		deviceEvent.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert failed: %w", err)
	}

	return nil
}
