package application

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"git.infra.egt.com/report/module/internal/domain"
)

type DeviceEventRepositoryInterface interface {
	FetchAll(ctx context.Context) ([]*domain.DeviceEvent, error)
	FetchByID(ctx context.Context, id *uuid.UUID) (*domain.DeviceEvent, error)
}

type DeviceEventFetcher struct {
	logger                *zap.Logger
	deviceEventRepository DeviceEventRepositoryInterface
}

func NewDeviceEventFetcher(logger *zap.Logger, deviceEventRepository DeviceEventRepositoryInterface) *DeviceEventFetcher {
	return &DeviceEventFetcher{
		logger:                logger,
		deviceEventRepository: deviceEventRepository,
	}
}

func (deviceEventFetcher *DeviceEventFetcher) GetById(ctx context.Context, id *uuid.UUID) (*domain.DeviceEvent, error) {
	deviceEvent, err := deviceEventFetcher.deviceEventRepository.FetchByID(ctx, id)
	if err != nil {
		deviceEventFetcher.logger.Error("fetch by id failed", zap.Error(err))

		return nil, err
	}

	return deviceEvent, nil
}

func (deviceEventFetcher *DeviceEventFetcher) GetAll(ctx context.Context) ([]*domain.DeviceEvent, error) {
	deviceEvents, err := deviceEventFetcher.deviceEventRepository.FetchAll(ctx)
	if err != nil {
		deviceEventFetcher.logger.Error("fetch by id failed", zap.Error(err))

		return nil, err
	}

	return deviceEvents, nil
}
