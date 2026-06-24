package domain

import (
	"time"

	"github.com/google/uuid"

	domainevent "git.infra.egt.com/gateway/module/internal/domain/event"
)

type DeviceEvent struct {
	ID         uuid.UUID
	DeviceId   string
	DeviceName string
	SensorName string
	Message    string
	CreatedAt  time.Time

	domainevent.DomainEventQueue
}

func NewDeviceEvent(
	id string,
	deviceId string,
	deviceName string,
	sensorName string,
	message string,
) (*DeviceEvent, error) {
	// TODO: Add domain logic here if needed
	// NOTE: this is simple validation and could be inside port, but I set it here as a domain rule - depends on the business requirements
	if err := uuid.Validate(id); err != nil {
		return nil, err
	}

	deviceEvent := &DeviceEvent{
		ID:         uuid.MustParse(id),
		DeviceId:   deviceId,
		DeviceName: deviceName,
		SensorName: sensorName,
		Message:    message,
		CreatedAt:  time.Now().UTC(),
	}

	domainDeviceEvent := &domainevent.DomainDeviceEvent{
		ID:         deviceEvent.ID,
		DeviceId:   deviceEvent.DeviceId,
		DeviceName: deviceEvent.DeviceName,
		SensorName: deviceEvent.SensorName,
		Message:    deviceEvent.Message,
		CreatedAt:  deviceEvent.CreatedAt,
	}

	// TODO: useful patern if there are more functions with domain logic and many places for enqueueing events
	deviceEvent.EnqueueEvent(&domainevent.DomainEvent{Event: domainDeviceEvent})

	return deviceEvent, nil
}
