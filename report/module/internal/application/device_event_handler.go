package application

import (
	"context"

	"git.infra.egt.com/report/module/internal/domain"
	"git.infra.egt.com/report/module/internal/dto"
	"go.uber.org/zap"
)

type DeviceEventSaverRepositoryInterface interface {
	Save(ctx context.Context, deviceEvent *domain.DeviceEvent) error
}

type DeviceEventHandler struct {
	logger                *zap.Logger
	deviceEventRepository DeviceEventSaverRepositoryInterface
}

func NewDeviceEventHandler(
	logger *zap.Logger,
	deviceEventRepository DeviceEventSaverRepositoryInterface,
) *DeviceEventHandler {
	return &DeviceEventHandler{
		logger:                logger,
		deviceEventRepository: deviceEventRepository,
	}
}

func (deviceEventHandler *DeviceEventHandler) Handle(ctx context.Context, deviceEventDTO *dto.DeviceEvent) error {
	deviceEvent, err := domain.NewDeviceEvent(
		deviceEventDTO.ID,
		deviceEventDTO.DeviceId,
		deviceEventDTO.DeviceName,
		deviceEventDTO.SensorName,
		deviceEventDTO.Message,
		deviceEventDTO.CreatedAt,
	)
	if err != nil {
		deviceEventHandler.logger.Error("deevice event handling failed", zap.Error(err))

		return err
	}

	err = deviceEventHandler.deviceEventRepository.Save(ctx, deviceEvent)
	if err != nil {
		deviceEventHandler.logger.Error("device event persisting failed", zap.Error(err))

		return err
	}

	return nil
}
