package kafka

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	"git.infra.egt.com/producer/module/infrastructure"
	"git.infra.egt.com/producer/module/infrastructure/database/repository"
)

type TransactionManagerInterface interface {
	WrapInTransaction(ctx context.Context, c func(ctx context.Context) error) error
}

type MessageRepositoryInterface interface {
	FetchBatch(
		ctx context.Context,
	) ([]*repository.Message, error)
	Delete(
		ctx context.Context,
		ids []int64,
	) error
}

type EventDispather struct {
	wg                 sync.WaitGroup
	cancel             context.CancelFunc
	transactionManager TransactionManagerInterface
	messageRepository  MessageRepositoryInterface
	logger             *zap.Logger
	deviceEventTopic   string
	kafkaAddresses     []string
}

func NewEventDispather(
	transactionManager TransactionManagerInterface,
	messageRepository MessageRepositoryInterface,
	logger *zap.Logger,
	c *infrastructure.Config,
) *EventDispather {
	ctx, cancel := context.WithCancel(context.Background())

	e := &EventDispather{
		transactionManager: transactionManager,
		cancel:             cancel,
		messageRepository:  messageRepository,
		logger:             logger,
		deviceEventTopic:   c.DeviceEventKafkaTopic,
		kafkaAddresses:     c.KafkaAddresses,
	}

	e.wg.Add(1)

	go e.dispatchDeviceEvents(ctx)

	return e
}

func (eventDispatcher *EventDispather) StopProcessingGracefully() {
	eventDispatcher.cancel()
	eventDispatcher.wg.Wait()
}

func (eventDispatcher *EventDispather) dispatchDeviceEvents(ctx context.Context) error {
	defer eventDispatcher.wg.Done()

	// TODO: kafka writer has to be a global and DI provided, not created here
	w := &kafka.Writer{
		Addr:                   kafka.TCP(eventDispatcher.kafkaAddresses...),
		AllowAutoTopicCreation: false,
	}

	defer func() {
		if err := w.Close(); err != nil {
			eventDispatcher.logger.Error("failed to close writer", zap.Error(err))
		}
	}()

	callable := func(ctx context.Context) error {
		msgs, err := eventDispatcher.messageRepository.FetchBatch(ctx)
		if err != nil {
			eventDispatcher.logger.Error("fetching failed", zap.Error(err))

			return err
		}

		if len(msgs) == 0 {
			time.Sleep(5 * time.Second)

			return nil
		}

		err = eventDispatcher.dispatch(ctx, w, msgs)
		if err != nil {
			eventDispatcher.logger.Error("diptaching failed", zap.Error(err))

			return err
		}

		var ids []int64
		for _, m := range msgs {
			ids = append(ids, m.ID)
		}

		err = eventDispatcher.messageRepository.Delete(ctx, ids)
		if err != nil {
			eventDispatcher.logger.Error("deliting failed", zap.Error(err))

			return err
		}

		return nil
	}

	for {
		select {
		case <-ctx.Done():
			eventDispatcher.logger.Info("Background worker received cancellation")

			return nil
		default:
			err := eventDispatcher.transactionManager.WrapInTransaction(ctx, callable)
			if err != nil {
				eventDispatcher.logger.Error("unable to disptach events", zap.Error(err))
			}
		}
	}
}

func (eventDispatcher *EventDispather) dispatch(
	ctx context.Context,
	kafkaWriter *kafka.Writer,
	msgs []*repository.Message,
) error {

	var messages []kafka.Message
	for _, m := range msgs {
		var headers []kafka.Header
		for kk, h := range m.Headers {
			headers = append(
				headers,
				kafka.Header{
					Key:   kk,
					Value: []byte(h),
				},
			)
		}

		key := uuid.New()
		messages = append(messages, kafka.Message{
			Key:     key[:],
			Headers: headers,
			Value:   m.Message,
			Topic:   eventDispatcher.deviceEventTopic,
		})
	}

	var err error
	const retries = 3
	for i := 0; i < retries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err = kafkaWriter.WriteMessages(ctx, messages...)
		if errors.Is(err, kafka.LeaderNotAvailable) || errors.Is(err, context.DeadlineExceeded) {
			time.Sleep(time.Millisecond * 250)
			continue
		}

		if err != nil {
			eventDispatcher.logger.Error("unexpected error", zap.Error(err))
		}
		break
	}

	eventDispatcher.logger.Info(fmt.Sprintf("dispatched %d messages", len(messages)))

	return err
}
