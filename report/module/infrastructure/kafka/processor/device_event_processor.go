package processor

import (
	"context"

	validatorV10 "github.com/go-playground/validator/v10"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"

	gatewaykafkav1 "git.infra.egt.com/clients/go/gateway/kafka/v1"
	"git.infra.egt.com/report/module/infrastructure"
	"git.infra.egt.com/report/module/internal/dto"
)

var validator = validatorV10.New()

type DeviceEventHandlerInterface interface {
	Handle(ctx context.Context, deviceEventDTO *dto.DeviceEvent) error
}

type DeviceEventProcessor struct {
	config             *infrastructure.Config
	logger             *zap.Logger
	deviceEventHandler DeviceEventHandlerInterface
}

func NewDeviceEventProcessor(config *infrastructure.Config, logger *zap.Logger, deviceEventHandler DeviceEventHandlerInterface) *DeviceEventProcessor {
	return &DeviceEventProcessor{
		config:             config,
		logger:             logger,
		deviceEventHandler: deviceEventHandler,
	}
}

func (deviceEventProcessor *DeviceEventProcessor) Process(
	ctx context.Context,
	kafkaMessage *kafka.Message,
) error {
	protoEvent, err := mapKafkaMessageToProto(kafkaMessage)
	if err != nil {
		deviceEventProcessor.logger.Error("invalid event structure", zap.Error(err))

		return nil
	}

	deviceEventDTO := mapProtoToDTO(protoEvent)
	if err := validateDeviceEventDTO(deviceEventDTO); err != nil {
		deviceEventProcessor.logger.Error("validation failed", zap.Error(err))

		return nil
	}

	err = deviceEventProcessor.deviceEventHandler.Handle(ctx, deviceEventDTO)
	if err != nil {
		deviceEventProcessor.logger.Error("device event consuming failed", zap.Error(err))

		return err
	}

	return nil
}

func mapKafkaMessageToProto(kafkaMessage *kafka.Message) (*gatewaykafkav1.Event, error) {
	messageBytes := kafkaMessage.Value

	event := &gatewaykafkav1.Event{}

	if err := protojson.Unmarshal(messageBytes, event); err != nil {
		return nil, err
	}

	return event, nil
}

func mapProtoToDTO(eventProto *gatewaykafkav1.Event) *dto.DeviceEvent {
	deviceEvent := &dto.DeviceEvent{}

	if eventProto.Id != "" {
		deviceEvent.ID = eventProto.Id
	}

	if eventProto.DeviceId != "" {
		deviceEvent.DeviceId = eventProto.DeviceId
	}

	if eventProto.DeviceName != "" {
		deviceEvent.DeviceName = eventProto.DeviceName
	}

	if eventProto.SensorName != "" {
		deviceEvent.SensorName = eventProto.SensorName
	}

	if eventProto.Message != "" {
		deviceEvent.Message = eventProto.Message
	}

	if eventProto.CreatedTime != nil {
		createdAt := eventProto.CreatedTime.AsTime()
		deviceEvent.CreatedAt = &createdAt
	}

	return deviceEvent
}

func validateDeviceEventDTO(deviceEventDTO *dto.DeviceEvent) error {
	if err := validator.Struct(deviceEventDTO); err != nil {
		return err
	}

	return nil
}
