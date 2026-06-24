package repository

import (
	"context"
	"fmt"

	"git.infra.egt.com/gateway/module/internal/domain"
)

type DeviceEventRepository struct {
	queryExecutor QueryExecutor
}

// NewDeviceEventRepository receives the pool from Fx.
func NewDeviceEventRepository(queryExecutor QueryExecutor) *DeviceEventRepository {
	return &DeviceEventRepository{queryExecutor: queryExecutor}
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
