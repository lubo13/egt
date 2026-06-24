package repository

import (
	"context"

	"go.uber.org/zap"

	"git.infra.egt.com/gateway/module/internal/domain"
	"git.infra.egt.com/gateway/module/internal/domain/event"
)

type DomainEventModelInterface interface {
	DequeueEvent() *event.DomainEvent
}

type DomainEventProducerInterface interface {
	ProduceEvent(ctx context.Context, domainEvent *event.DomainEvent) error
}

type ModelRepositoryInterface interface {
	Save(ctx context.Context, deviceEvent *domain.DeviceEvent) error
}

type DomainEventRepositoryDecorator struct {
	logger              *zap.Logger
	modelRepository     ModelRepositoryInterface
	domainEventProducer DomainEventProducerInterface
}

func NewDomainEventRepositoryDecorator(
	logger *zap.Logger,
	modelRepository ModelRepositoryInterface,
	domainEventProducer DomainEventProducerInterface,
) *DomainEventRepositoryDecorator {
	return &DomainEventRepositoryDecorator{
		logger:              logger,
		modelRepository:     modelRepository,
		domainEventProducer: domainEventProducer,
	}
}

func (r *DomainEventRepositoryDecorator) Save(ctx context.Context, deviceEvent *domain.DeviceEvent) error {
	if _, ok := interface{}(deviceEvent).(DomainEventModelInterface); ok {
		for {
			event := deviceEvent.DequeueEvent()
			if event == nil {
				break
			}

			if err := r.domainEventProducer.ProduceEvent(ctx, event); err != nil {
				r.logger.Error("Failed to handle event", zap.Error(err))
				return err
			}
		}
	}

	return r.modelRepository.Save(ctx, deviceEvent)
}
