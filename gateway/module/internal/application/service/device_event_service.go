package service

import (
	"context"
	"sync"

	"go.uber.org/zap"

	"git.infra.egt.com/gateway/module/internal/domain"
	"git.infra.egt.com/gateway/module/internal/dto"
)

const GOROUTINES_NUM = 5
const BUFFER_CAP = 25

type TransactionManagerInterface interface {
	WrapInTransaction(ctx context.Context, c func(ctx context.Context) error) error
}

type DeviceEventRepositoryInterface interface {
	Save(ctx context.Context, deviceEvent *domain.DeviceEvent) error
}

// DeviceEventService handle events processing.
type DeviceEventService struct {
	wg                    sync.WaitGroup
	cancel                context.CancelFunc
	log                   *zap.Logger
	deviceEventRepository DeviceEventRepositoryInterface
	transactionManager    TransactionManagerInterface
	eventChannel          chan *dto.DeviceEvent
}

func NewDeviceEventService(log *zap.Logger,
	deviceEventRepository DeviceEventRepositoryInterface,
	transactionManager TransactionManagerInterface,
) *DeviceEventService {
	ctx, cancel := context.WithCancel(context.Background())

	deviceeventService := &DeviceEventService{
		cancel:                cancel,
		log:                   log,
		deviceEventRepository: deviceEventRepository,
		transactionManager:    transactionManager,
		eventChannel:          make(chan *dto.DeviceEvent, BUFFER_CAP),
	}

	for i := 0; i < GOROUTINES_NUM; i++ {
		deviceeventService.wg.Add(1)

		go deviceeventService.processing(ctx)
	}

	return deviceeventService
}

func (s *DeviceEventService) Add(deviceEventDTO *dto.DeviceEvent) {
	s.eventChannel <- deviceEventDTO
}

func (s *DeviceEventService) StopProcessingGracefully() {
	s.cancel()
	s.wg.Wait()
	close(s.eventChannel)
}

func (s *DeviceEventService) processing(ctx context.Context) error {
	defer s.wg.Done()

	eventProcessing := func(ctx context.Context, e *dto.DeviceEvent) {
		s.log.Info("Write to Database and outbox table")

		deviceEventModel, err := domain.NewDeviceEvent(
			e.ID,
			e.DeviceId,
			e.DeviceName,
			e.SensorName,
			e.Message,
		)
		if err != nil {
			s.log.Error("domain validation failed ", zap.Error(err))

			return
		}

		callable := func(ctx context.Context) error {
			return s.deviceEventRepository.Save(ctx, deviceEventModel)
		}

		err = s.transactionManager.WrapInTransaction(ctx, callable)
		if err != nil {
			s.log.Error("persist failed", zap.Error(err))
		}
	}

	for {
		select {
		case <-ctx.Done():
			s.log.Info("Background worker received cancellation")

			return nil
		case e := <-s.eventChannel:
			eventProcessing(ctx, e)
		}
	}
}
