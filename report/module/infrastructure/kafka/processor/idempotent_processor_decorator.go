package processor

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	"git.infra.egt.com/report/module/infrastructure"
)

const X_IDEMOPOTENT_ID string = "X-Idempotent-Id"

type IdempotentIdRepositoryInterface interface {
	IsExists(ctx context.Context, id *uuid.UUID) (bool, error)
	Save(ctx context.Context, id *uuid.UUID) error
}

type TransactionManagerInterface interface {
	WrapInTransaction(ctx context.Context, c func(ctx context.Context) error) error
}

type IdempotentProcessorDecorator struct {
	config                 *infrastructure.Config
	logger                 *zap.Logger
	idempotentIdRepository IdempotentIdRepositoryInterface
	messangeProcessor      MessageProcessorInterface
	transactionManager     TransactionManagerInterface
}

func NewIdempotentProcessorDecorator(
	config *infrastructure.Config,
	logger *zap.Logger,
	idempotentIdRepository IdempotentIdRepositoryInterface,
	messangeProcessor MessageProcessorInterface,
	transactionManager TransactionManagerInterface,
) *IdempotentProcessorDecorator {
	return &IdempotentProcessorDecorator{
		config:                 config,
		logger:                 logger,
		idempotentIdRepository: idempotentIdRepository,
		messangeProcessor:      messangeProcessor,
		transactionManager:     transactionManager,
	}
}

func (idempotentProcessorDecorator *IdempotentProcessorDecorator) Process(
	ctx context.Context,
	kafkaMessage *kafka.Message,
) error {
	callable := func(ctx context.Context) error {
		var idempotentId uuid.UUID
		for _, header := range kafkaMessage.Headers {
			if header.Key == X_IDEMOPOTENT_ID {
				if err := uuid.Validate(string(header.Value)); err != nil {
					idempotentProcessorDecorator.logger.Error("invalid idempotent id", zap.Error(err))

					return nil
				}

				idempotentId = uuid.MustParse(string(header.Value))
			}
		}

		if idempotentId == uuid.Nil {
			idempotentProcessorDecorator.logger.Error("missing idempotent id")
		}

		isExist, err := idempotentProcessorDecorator.idempotentIdRepository.IsExists(ctx, &idempotentId)
		if err != nil {
			idempotentProcessorDecorator.logger.Error("idempotent key checking failed", zap.Error(err))

			return err
		}
		if isExist {
			idempotentProcessorDecorator.logger.Info(fmt.Sprintf("skip duplicated id -> %s", idempotentId.String()))

			return nil
		}

		err = idempotentProcessorDecorator.messangeProcessor.Process(ctx, kafkaMessage)
		if err != nil {
			idempotentProcessorDecorator.logger.Error("processing failed", zap.Error(err))

			return err
		}

		if err := idempotentProcessorDecorator.idempotentIdRepository.Save(ctx, &idempotentId); err != nil {
			idempotentProcessorDecorator.logger.Error("saving of idempotent id failed", zap.Error(err))

			return err
		}

		return nil
	}

	return idempotentProcessorDecorator.transactionManager.WrapInTransaction(
		ctx,
		callable,
	)
}
