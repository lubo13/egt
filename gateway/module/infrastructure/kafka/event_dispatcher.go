package kafka

import (
	"context"
	"encoding/json"
	"time"

	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"

	gatewaykafkav1 "git.infra.egt.com/clients/go/gateway/kafka/v1"
	"git.infra.egt.com/gateway/module/infrastructure"
	"git.infra.egt.com/gateway/module/internal/domain/event"
)

const X_IDEMOPOTENT_ID string = "X-Idempotent-Id"

type MessageRepositoryInterface interface {
	Produce(ctx context.Context, topic string, headers []byte, message []byte, createdAt time.Time) error
}

type EventDispather struct {
	messageRepository MessageRepositoryInterface
	logger            *zap.Logger
	deviceEventTopic  string
}

func NewEventDispather(
	messageRepository MessageRepositoryInterface,
	logger *zap.Logger,
	c *infrastructure.Config,
) *EventDispather {
	return &EventDispather{
		messageRepository: messageRepository,
		logger:            logger,
		deviceEventTopic:  c.DeviceEventKafkaTopic,
	}
}

func (eventDispatcher *EventDispather) ProduceEvent(ctx context.Context, domainEvent *event.DomainEvent) error {
	topic, headersJson, messageJson := eventDispatcher.messageMapper(domainEvent)
	if topic == "" || headersJson == nil || messageJson == nil {
		return nil
	}

	return eventDispatcher.messageRepository.Produce(ctx, topic, headersJson, messageJson, time.Now().UTC())
}

func (eventDispatcher *EventDispather) messageMapper(domainEvent *event.DomainEvent) (string, []byte, []byte) {
	e := domainEvent.Event

	switch ev := e.(type) {
	case *event.DomainDeviceEvent:
		protoEvent := &gatewaykafkav1.Event{
			Id:          ev.ID.String(),
			DeviceId:    ev.DeviceId,
			DeviceName:  ev.DeviceName,
			SensorName:  ev.SensorName,
			Message:     ev.Message,
			CreatedTime: timestamppb.New(ev.CreatedAt),
		}

		messageJson := protojson.Format(protoEvent)
		headers := &map[string]string{
			X_IDEMOPOTENT_ID: ev.ID.String(),
		}
		headersJson, _ := json.Marshal(headers)

		eventDispatcher.logger.Info("Successfully mapped")

		return eventDispatcher.deviceEventTopic, headersJson, []byte(messageJson)
	default:
		return "", []byte{}, []byte{}
	}
}
