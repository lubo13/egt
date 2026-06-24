package main

import (
	"context"
	"errors"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	"git.infra.egt.com/report/module/infrastructure"
	"git.infra.egt.com/report/module/infrastructure/kafka/processor"
)

type DeviceEventConsumer struct {
	config               *infrastructure.Config
	logger               *zap.Logger
	deviceEventProcessor processor.MessageProcessorInterface
}

func NewDeviceEventConsumer(
	config *infrastructure.Config,
	logger *zap.Logger,
	deviceEventProcessor processor.MessageProcessorInterface,
) *DeviceEventConsumer {
	return &DeviceEventConsumer{
		config:               config,
		logger:               logger,
		deviceEventProcessor: deviceEventProcessor,
	}
}

func (deviceEventConsumer *DeviceEventConsumer) Consume(ctx context.Context, worker *Wroker) {
	defer worker.Done()

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:          deviceEventConsumer.config.KafkaAddresses,
		GroupID:          deviceEventConsumer.config.ConsumerGroup,
		Topic:            deviceEventConsumer.config.DeviceEventKafkaTopic,
		MaxBytes:         10e6, // 10MB,
		RebalanceTimeout: time.Duration(5 * time.Second),
	})

	deviceEventConsumer.logger.Info("device event consumer start")

outer:
	for {
		select {
		case <-ctx.Done():
			deviceEventConsumer.logger.Info("background worker received cancellation")

			break outer
		default:
			m, err := r.FetchMessage(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					deviceEventConsumer.logger.Info("context cancelled, stopping consumer")

					break outer
				}

				deviceEventConsumer.logger.Error("failed to fetch message", zap.Error(err))
			}

			// In case of transient error retry with a delay
			for i := 1; i < 3; i++ {
				if err := deviceEventConsumer.deviceEventProcessor.Process(ctx, &m); err != nil {
					time.Sleep(time.Duration(i) * time.Second)

					continue
				}

				break
			}

			if err := r.CommitMessages(ctx, m); err != nil {
				deviceEventConsumer.logger.Error("failed to commit messages", zap.Error(err))
			}
		}

	}

	if err := r.Close(); err != nil {
		deviceEventConsumer.logger.Error("failed to close reader", zap.Error(err))
	}
}
